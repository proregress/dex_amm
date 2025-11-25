package block

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/common"
	"github.com/blocto/solana-go-sdk/program/token"
	"github.com/blocto/solana-go-sdk/rpc"
	solTypes "github.com/blocto/solana-go-sdk/types"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/gorilla/websocket"
	"github.com/mr-tron/base58"
	"github.com/panjf2000/ants/v2"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/threading"
	"richcode.cc/dex/consumer/internal/config"
	"richcode.cc/dex/consumer/internal/svc"
	"richcode.cc/dex/model/solmodel"
	"richcode.cc/dex/pkg/constants"
	"richcode.cc/dex/pkg/types"
)

var ErrServiceStop = errors.New("service stop")

var ErrUnknowProgram = errors.New("unknow program")

type BlockService struct {
	Name string
	sc   *svc.ServiceContext
	/* Solana Go SDK 的客户端，用于与 Solana 区块链节点通信。
	Solana Go SDK 是 Solana 区块链的官方 Go 语言 SDK，提供了与 Solana 网络交互的工具和库。
	它包含了与 Solana 区块链节点通信的各种功能，包括获取区块数据、发送交易、查询账户信息等。
	*/
	c *client.Client
	logx.Logger
	workerPool  *ants.Pool
	ctx         context.Context
	cancel      func(err error)
	slotChannel chan uint64
	slot        uint64
	Conn        *websocket.Conn
	solPrice    float64
	name        string
}

// Start implements service.Service.
func (s *BlockService) Start() {
	s.GetBlockFromHttp()
}

// Stop implements service.Service.
func (s *BlockService) Stop() {
	s.cancel(ErrServiceStop)
	if s.Conn != nil {
		/* 通过 WebSocket 发送 JSON-RPC 消息，调用 blockUnsubscribe 取消订阅（订阅 ID 为 0）
		-服务停止前需要取消订阅，避免资源泄漏
		-通知 Solana 节点停止推送区块更新
		-确保 WebSocket 连接正确关闭
		*/
		err := s.Conn.WriteMessage(websocket.TextMessage, []byte("{\"id\":1,\"jsonrpc\":\"2.0\",\"method\": \"blockUnsubscribe\", \"params\": [0]}\n"))
		if err != nil {
			s.Error("programUnsubscribe", err)
		}
		_ = s.Conn.Close()
	}
}

func NewBlockService(sc *svc.ServiceContext, name string, slotChnanel chan uint64, index int) *BlockService {
	ctx, cancel := context.WithCancelCause(context.Background())
	pool, _ := ants.NewPool(5)
	blockService := &BlockService{
		c: client.New(rpc.WithEndpoint(config.FindChainRpcByChainId(constants.SolChainIdInt)), rpc.WithHTTPClient(&http.Client{
			Timeout: 5 * time.Second,
		})),
		sc:          sc,
		name:        name,
		workerPool:  pool,
		slotChannel: slotChnanel,
		Logger:      logx.WithContext(context.Background()).WithFields(logx.Field("service", fmt.Sprintf("%s-%v", name, index))),
		ctx:         ctx,
		cancel:      cancel,
	}
	return blockService
}

/*基于HTTP接口，通过传入slot编号，获取区块数据*/
func (s *BlockService) GetBlockFromHttp() {
	fmt.Print("block service is started")
	ctx := s.ctx
	for {
		select {
		case <-s.ctx.Done():
			fmt.Print("block service is stopped")
			return
		case slot, ok := <-s.slotChannel: // 异步地从channel中获取slot数据
			if !ok {
				fmt.Print("slotChan is closed")
				return
			}
			//打印当前最新slot
			fmt.Println("consume slot is:", slot)
			// 再创建一个单独的协程，专门用来处理区块数据
			// RunSafe：异步执行，不会阻塞当前协程，继续执行下一个
			// RunSafe和GoSafe的区别？？？
			threading.RunSafe(func() {
				s.ProcessBlock(ctx, int64(slot)) // 直接调用的话是同步执行，会阻塞在这一步，直到处理完才会继续下一个
			})
		}
	}
}

func (s *BlockService) FillTradeWithPairInfo(trade *types.TradeWithPair, slot int64) {
	trade.Slot = slot
	trade.BlockNum = slot
	trade.ChainIdInt = constants.SolChainIdInt
	trade.ChainId = constants.SolChainId
}

func (s *BlockService) ProcessBlock(ctx context.Context, slot int64) {
	if slot == 0 {
		return
	}

	// Step0: 初始化区块对象，并将状态默认标记为失败，后续流程成功再回写
	beginTime := time.Now()

	s.slot = uint64(slot)

	// 创建block对象
	block := &solmodel.Block{
		Slot:   slot,
		Status: constants.BlockFailed,
	}

	blockInfo, err := GetSolBlockInfoDelay(s.sc.GetSolClient(), ctx, uint64(slot))
	if err != nil || blockInfo == nil {
		fmt.Println("get block info error: ", err)

		if strings.Contains(err.Error(), "was skipped") { // 区块里没有任何有意义的交易（比如只有一些打包信息投票信息啥的）
			block.Status = constants.BlockSkipped
		}

		_ = s.sc.BlockModel.Insert(ctx, block)
		return
	}
	// Step1: 从上面拿到的blockInfo将信息设置进block对象中
	if blockInfo.BlockTime != nil {
		block.BlockTime = *blockInfo.BlockTime
	}

	if blockInfo.BlockHeight != nil {
		block.BlockHeight = *blockInfo.BlockHeight
	}
	block.Status = constants.BlockProcessed

	// Step2: 计算当期 SOL 价格，为后续成交估值提供基准
	var tokenAccountMap = make(map[string]*TokenAccount) // string：ATA账户地址，map：通过ATA账户地址拿到完整的账户对象
	solPrice := s.GetBlockSolPrice(ctx, blockInfo, tokenAccountMap)
	if solPrice == 0 {
		solPrice = s.solPrice
	}
	// 区块 -> 交易 -> 转账（Transfer）SOL-(USDT/USDC) -> (USDT|USDC) / SOL = 价格
	block.SolPrice = solPrice

	// 初始化一个存放交易信息的切片，初始容量设为1000
	// 后续会把从区块中解析出来的交易(TradeWithPair)加入到这个切片中，便于统一处理（如分类落库等）
	trades := make([]*types.TradeWithPair, 0, 1000)

	// 通过slice组件遍历区块中的每一笔链上交易，进行处理，即遍历transactions，拿到每一个交易对象tx
	slice.ForEach(blockInfo.Transactions, func(index int, tx client.BlockTransaction) {
		// Step3: 构造解码上下文。这里新建一个 DecodedTx 结构体用来保存当前交易的相关信息：
		// BlockDb            当前处理的区块指针
		// Tx                 当前遍历到的链上交易指针
		// TxIndex            交易在区块内的序号
		// TokenAccountMap    本次区块处理中维护的token账户（复用，提升效率）
		decodeTx := &DecodedTx{
			BlockDb:         block,
			Tx:              &tx,
			TxIndex:         index,
			TokenAccountMap: tokenAccountMap,
		}

		// 解码链上交易，返回解码后的 TradeWithPair 切片
		trade, err := DecodeTx(ctx, s.sc, decodeTx)
		if err != nil {
			// 如果是未识别的合约（unknow program），则直接跳过此条
			if strings.Contains(err.Error(), "unknow program") {
				return
			}
			// 其他解码出错则输出日志并跳过
			fmt.Println("decode tx failed: ", err.Error())
			return
		}

		// 对解码出来的 trade 结果再次过滤，只保留非空项；
		// 并填充 TradeWithPair 的补充信息（如引用了Slot等元数据）
		trade = slice.Filter(trade, func(index int, item *types.TradeWithPair) bool {
			if item == nil {
				return false
			}
			s.FillTradeWithPairInfo(item, slot)
			return true
		})

		// 将本笔交易对应的所有成交记入 trades 切片，后面统一处理
		trades = append(trades, trade...)
	})

	// Step4: 将成交按 Pair 归类，方便后续批量写入
	tradeMap := make(map[string][]*types.TradeWithPair)

	pumpSwapCount := 0
	pumpFunCount := 0

	for _, trade := range trades {
		if len(trade.PairAddr) > 0 {
			tradeMap[trade.PairAddr] = append(tradeMap[trade.PairAddr], trade)
		}
	}

	for _, value := range tradeMap {
		// 简单统计不同 Swap 的成交数量，方便监控
		if value[0].SwapName == constants.PumpFun {
			pumpFunCount++
			continue
		}
		if value[0].SwapName == constants.PumpSwap {
			pumpSwapCount++
			continue
		}
	}

	{
		// 额外挑出 Mint 行为，触发 Token 总量刷新
		tokenMints := slice.Filter[*types.TradeWithPair](trades, func(_ int, item *types.TradeWithPair) bool {
			if item != nil && item.Type == types.TradeTokenMint {
				return true
			}
			return false
		})

		s.UpdateTokenMints(ctx, tokenMints)
		s.Infof("processBlock:%v UpdateTokenMints size: %v, dur: %v, tokenMints: %v", slot, len(tokenMints), time.Since(beginTime), len(tokenMints))
	}

	{
		// 额外挑出 Burn 行为，触发 Token 总量刷新
		tokenBurns := slice.Filter[*types.TradeWithPair](trades, func(_ int, item *types.TradeWithPair) bool {
			if item != nil && item.Type == types.TradeTokenBurn {
				return true
			}
			return false
		})

		s.UpdateTokenBurns(ctx, tokenBurns)
		s.Infof("processBlock:%v UpdateTokenBurns size: %v, dur: %v, tokenBurns: %v", slot, len(tokenBurns), time.Since(beginTime), len(tokenBurns))
	}

	//并发处理： 保存交易信息，保存token账户信息
	group := threading.NewRoutineGroup()
	group.RunSafe(func() {
		// Step5: 写入成交信息 & TokenAccount 快照
		s.SaveTrades(ctx, constants.SolChainIdInt, tradeMap)
		s.Infof("processBlock:%v saveTrades tx_size: %v, dur: %v, trade_size: %v", slot, len(blockInfo.Transactions), time.Since(beginTime), len(trades))

		s.SaveTokenAccounts(ctx, trades, tokenAccountMap)
		s.Infof("processBlock:%v saveTokenAccounts tx_size: %v, dur: %v, trade_size: %v", slot, len(blockInfo.Transactions), time.Since(beginTime), len(trades))
	})

	// pump swap
	group.RunSafe(func() {
		// Step6: 针对 PumpSwap 交易补充池子元数据
		slice.ForEach(trades, func(_ int, trade *types.TradeWithPair) {
			if trade.SwapName == constants.PumpSwap || trade.SwapName == "PumpSwap" {
				if trade.Type == types.TradeTypeBuy || trade.Type == types.TradeTypeSell {
					if err = s.SavePumpSwapPoolInfo(ctx, trade); err != nil {
						s.Errorf("processBlock:%v SavePumpSwapPoolInfo err: %v", slot, err)
					}
				}

			}
		})
	})

	group.Wait()

	// Step7: 区块落库，标识处理完成
	err = s.sc.BlockModel.Insert(ctx, block)
	if err != nil {
		s.Error("insert block error", err)
	}
}

func DecodeTx(ctx context.Context, sc *svc.ServiceContext, dtx *DecodedTx) (trades []*types.TradeWithPair, err error) {
	if dtx.Tx == nil || dtx.BlockDb == nil {
		return
	}

	tx := dtx.Tx
	dtx.TxHash = base58.Encode(tx.Transaction.Signatures[0])

	if tx.Meta.Err != nil {
		return
	}

	dtx.InnerInstructionMap = GetInnerInstructionMap(tx)

	// 遍历交易内所有指令，挨个解析生成业务侧的成交结构
	// Instructions是交易中的所有指令，遍历每一个指令inst
	for i := range tx.Transaction.Message.Instructions {
		inst := &tx.Transaction.Message.Instructions[i]
		var trade *types.TradeWithPair // tradewithpair ：对应交易对的数据，即池子的数据
		trade, err = DecodeInstruction(ctx, sc, dtx, inst, i)
		if err != nil {
			return nil, err
		}
		trades = append(trades, trade)
	}
	return
}

func DecodeInstruction(ctx context.Context, sc *svc.ServiceContext, dtx *DecodedTx, instruction *solTypes.CompiledInstruction, index int) (trade *types.TradeWithPair, err error) {
	if len(dtx.Tx.AccountKeys) == 0 {
		return nil, errors.New("account keys is empty")
	}

	if int(instruction.ProgramIDIndex) >= len(dtx.Tx.AccountKeys) {
		return nil, fmt.Errorf("program ID index %d out of bounds for account keys length %d", instruction.ProgramIDIndex, len(dtx.Tx.AccountKeys))
	}

	tx := dtx.Tx
	// AccountKeys是当前交易中的所有指令的账户id
	// ProgramIDIndex是当前指令的账户id索引
	program := tx.AccountKeys[instruction.ProgramIDIndex].String()

	switch program {
	case ProgramStrPumpFun:
		trade, err = DecodePumpFunInstruction(instruction, tx)
		return
	case ProgramStrPumpFunAMM:
		decoder := &PumpAmmDecoder{
			ctx:                 ctx,
			svcCtx:              sc,
			dtx:                 dtx,
			compiledInstruction: instruction,
		}
		trade, err = decoder.DecodePumpFunAMMInstruction()
		return
	default:
		return nil, ErrUnknowProgram
	}
}

func GetInnerInstructionByInner(instructions []solTypes.CompiledInstruction, startIndex, innerLen int) *client.InnerInstruction {
	if startIndex+innerLen+1 > len(instructions) {
		return nil
	}
	innerInstruction := &client.InnerInstruction{
		Index: uint64(instructions[startIndex].ProgramIDIndex),
	}
	for i := 0; i < innerLen; i++ {
		innerInstruction.Instructions = append(innerInstruction.Instructions, instructions[startIndex+i+1])
	}
	return innerInstruction
}

// FillTokenAccountMap 填充交易中的tokenAccount 数据
func FillTokenAccountMap(tx *client.BlockTransaction, tokenAccountMapIn map[string]*TokenAccount) (tokenAccountMap map[string]*TokenAccount, hasTokenChange bool) {
	if tokenAccountMapIn == nil {
		tokenAccountMapIn = make(map[string]*TokenAccount)
	}
	tokenAccountMap = tokenAccountMapIn
	// 遍历执行交易前各个代币账户余额  	PreTokenBalances：交易前各个代币账户余额
	for _, pre := range tx.Meta.PreTokenBalances {
		var tokenAccount = tx.AccountKeys[pre.AccountIndex].String()
		preValue, _ := strconv.ParseInt(pre.UITokenAmount.Amount, 10, 64)
		tokenAccountMap[tokenAccount] = &TokenAccount{
			Owner:               pre.Owner,                  // owner address
			TokenAccountAddress: tokenAccount,               // token account address
			TokenAddress:        pre.Mint,                   // token address
			TokenDecimal:        pre.UITokenAmount.Decimals, // token decimal
			PreValue:            preValue,
			Closed:              true,
			PreValueUIString:    pre.UITokenAmount.UIAmountString,
		}
	}
	// 遍历执行交易后各个代币账户余额  	PostTokenBalances：交易后各个代币账户余额
	for _, post := range tx.Meta.PostTokenBalances {
		var tokenAccount = tx.AccountKeys[post.AccountIndex].String()
		postValue, _ := strconv.ParseInt(post.UITokenAmount.Amount, 10, 64)
		if tokenAccountMap[tokenAccount] != nil {
			tokenAccountMap[tokenAccount].Closed = false
			tokenAccountMap[tokenAccount].PostValue = postValue
			if tokenAccountMap[tokenAccount].PostValue != tokenAccountMap[tokenAccount].PreValue {
				hasTokenChange = true
			}
		} else {
			hasTokenChange = true
			tokenAccountMap[tokenAccount] = &TokenAccount{
				Owner:               post.Owner,                  // owner address
				TokenAccountAddress: tokenAccount,                // token account address
				TokenAddress:        post.Mint,                   // token address
				TokenDecimal:        post.UITokenAmount.Decimals, // token decimal
				PostValue:           postValue,                   // token balance
				Init:                true,
				PostValueUIString:   post.UITokenAmount.UIAmountString,
			}
		}
	}

	// 遍历交易中的主指令（Instructions），查找Token程序的初始化账户指令
	// 这些指令用于创建新的代币账户，需要解析并添加到tokenAccountMap中
	for i := range tx.Transaction.Message.Instructions {
		instruction := &tx.Transaction.Message.Instructions[i]
		program := tx.AccountKeys[instruction.ProgramIDIndex].String()
		if program == ProgramStrToken {
			DecodeInitAccountInstruction(tx, tokenAccountMap, instruction)
		}
	}
	// 遍历交易中的内部指令（InnerInstructions），查找Token程序的初始化账户指令
	// InnerInstructions是主指令执行过程中产生的嵌套指令，也可能包含账户初始化操作
	for _, instructions := range tx.Meta.InnerInstructions {
		for i := range instructions.Instructions {
			instruction := instructions.Instructions[i]
			program := tx.AccountKeys[instruction.ProgramIDIndex].String()
			if program == ProgramStrToken {
				DecodeInitAccountInstruction(tx, tokenAccountMap, &instruction)
			}
		}
	}
	// 创建tokenDecimalMap，用于存储已知的代币精度信息
	// key为代币地址（TokenAddress），value为代币精度（TokenDecimal）
	tokenDecimalMap := make(map[string]uint8)
	// 遍历tokenAccountMap，收集所有已知的代币精度（不为0的）
	// 这样可以建立一个代币地址到精度的映射关系
	for _, v := range tokenAccountMap {
		if v.TokenDecimal != 0 {
			tokenDecimalMap[v.TokenAddress] = v.TokenDecimal
		}
	}
	// 对于tokenAccountMap中精度为0的账户，尝试从tokenDecimalMap中查找并填充对应的精度值
	// 这样可以确保同一个代币的所有账户都使用相同的精度值
	for _, v := range tokenAccountMap {
		if v.TokenDecimal == 0 {
			v.TokenDecimal = tokenDecimalMap[v.TokenAddress]
		}
	}
	return
}

// DecodeInitAccountInstruction 解码 Solana Token 程序的初始化账户指令
// 该函数用于解析交易中创建新代币账户的指令，并将账户信息添加到 tokenAccountMap 中
// 参数说明：
//   - tx: 区块交易对象，包含账户密钥和交易信息
//   - tokenAccountMap: 代币账户映射表，key 为代币账户地址，value 为代币账户详细信息
//   - instruction: 编译后的指令对象，包含指令类型、账户索引和数据
func DecodeInitAccountInstruction(tx *client.BlockTransaction, tokenAccountMap map[string]*TokenAccount, instruction *solTypes.CompiledInstruction) {
	// 如果指令数据为空，无法解析，直接返回
	if len(instruction.Data) == 0 {
		return
	}
	var mint, tokenAccount, owner string
	// 根据指令数据的第一个字节判断指令类型
	switch token.Instruction(instruction.Data[0]) {
	// InitializeAccount: 最基础的初始化代币账户指令
	// 业务含义：在 Solana 网络上创建一个新的代币账户（Associated Token Account, ATA）
	// 该账户用于存储特定代币（mint）的余额，并关联到指定的所有者（owner）
	// 账户结构：需要 3 个账户参数
	//   - Accounts[0]: 要初始化的代币账户地址（Token Account）
	//   - Accounts[1]: 代币的 mint 地址（Token Mint Address）
	//   - Accounts[2]: 账户所有者地址（Owner Address）
	case token.InstructionInitializeAccount:
		// 验证账户参数数量是否足够
		if len(instruction.Accounts) < 3 {
			return
		}
		// 从账户列表中提取代币账户地址、mint 地址和所有者地址
		tokenAccount = tx.AccountKeys[instruction.Accounts[0]].String()
		mint = tx.AccountKeys[instruction.Accounts[1]].String()
		owner = tx.AccountKeys[instruction.Accounts[2]].String()

	// InitializeAccount2: 改进版的初始化账户指令
	// 业务含义：与 InitializeAccount 功能相同，但优化了参数传递方式
	// 主要区别：owner 地址不再作为账户参数传递，而是编码在指令数据（Data）中
	// 优势：减少账户参数数量，降低交易成本，提高效率
	// 账户结构：只需要 2 个账户参数
	//   - Accounts[0]: 要初始化的代币账户地址（Token Account）
	//   - Accounts[1]: 代币的 mint 地址（Token Mint Address）
	// 数据格式：Data[1:33] 包含 32 字节的 owner 公钥
	case token.InstructionInitializeAccount2:
		// 验证账户参数数量和数据长度是否足够
		// 需要至少 2 个账户参数，数据长度至少 33 字节（1 字节指令类型 + 32 字节 owner 公钥）
		if len(instruction.Accounts) < 2 || len(instruction.Data) < 33 {
			return
		}
		// 从账户列表中提取代币账户地址和 mint 地址
		tokenAccount = tx.AccountKeys[instruction.Accounts[0]].String()
		mint = tx.AccountKeys[instruction.Accounts[1]].String()
		// 从指令数据中解析 owner 地址（跳过第一个字节的指令类型）
		owner = common.PublicKeyFromBytes(instruction.Data[1:]).String()

	// InitializeAccount3: 最新版本的初始化账户指令
	// 业务含义：与 InitializeAccount2 功能相同，是 Solana Token 程序的最新实现
	// 主要区别：进一步优化了指令格式，owner 地址同样编码在指令数据中
	// 使用场景：新创建的交易通常使用此版本，以获得更好的性能和成本效益
	// 账户结构：只需要 2 个账户参数
	//   - Accounts[0]: 要初始化的代币账户地址（Token Account）
	//   - Accounts[1]: 代币的 mint 地址（Token Mint Address）
	// 数据格式：Data[1:33] 包含 32 字节的 owner 公钥
	case token.InstructionInitializeAccount3:
		// 验证账户参数数量和数据长度是否足够
		// 需要至少 2 个账户参数，数据长度至少 33 字节（1 字节指令类型 + 32 字节 owner 公钥）
		if len(instruction.Accounts) < 2 || len(instruction.Data) < 33 {
			return
		}
		// 从账户列表中提取代币账户地址和 mint 地址
		tokenAccount = tx.AccountKeys[instruction.Accounts[0]].String()
		mint = tx.AccountKeys[instruction.Accounts[1]].String()
		// 从指令数据中解析 owner 地址（跳过第一个字节的指令类型）
		owner = common.PublicKeyFromBytes(instruction.Data[1:]).String()

	// 其他未知的指令类型，无法处理，直接返回
	default:
		return
	}
	// 检查该代币账户是否已存在于映射表中，且 mint 地址匹配
	// 如果已存在且 mint 地址相同，说明账户信息已记录，无需重复添加
	if tokenAccountMap[tokenAccount] != nil && tokenAccountMap[tokenAccount].TokenAddress == mint {
		return
	} else {
		// 创建新的代币账户记录并添加到映射表中
		// 标记账户已初始化，设置所有者、代币地址和账户地址
		// 初始余额和精度暂时设为 0，后续会从交易元数据中填充
		tokenAccountMap[tokenAccount] = &TokenAccount{
			Init:                true,         // 标记账户已初始化
			Owner:               owner,        // 账户所有者地址
			TokenAddress:        mint,         // 代币 mint 地址
			TokenAccountAddress: tokenAccount, // 代币账户地址
			TokenDecimal:        0,            // 代币精度（初始为 0，后续会填充）
			PreValue:            0,            // 交易前余额（初始为 0）
			PostValue:           0,            // 交易后余额（初始为 0）
		}
	}
}

// DecodeTokenTransfer 解析transfer指令
func DecodeTokenTransfer(accountKeys []common.PublicKey, instruction *solTypes.CompiledInstruction) (transfer *token.TransferParam, err error) {
	transfer = &token.TransferParam{}
	// 解析transfer指令
	if accountKeys[instruction.ProgramIDIndex].String() == common.Token2022ProgramID.String() { // Token2022ProgramID
		// Instruction Accounts 参与指令的账户在交易账户列表中的索引，source、destination、authority 三个账户
		if len(instruction.Accounts) < 3 {
			err = errors.New("not enough accounts")
			return
		}
		// Instruction Data ：指令的二进制数据（传给程序的具体参数arguments） 两个参数，discriminator描述符、Amount数量
		if len(instruction.Data) < 1 {
			err = errors.New("data len too small")
			return
		}
		// 第一个参数discriminator描述符 判断指令类型是 transfer 还是 transferChecked(3和12 只是索引值)
		if instruction.Data[0] == byte(token.InstructionTransfer) {
			// transfer指令长度一共9个字节
			if len(instruction.Data) != 9 {
				err = errors.New("data len not equal 9")
				return
			}
			if len(instruction.Accounts) < 3 {
				err = errors.New("account len too small")
				return
			}
			transfer.From = accountKeys[instruction.Accounts[0]]                // 发送方账户
			transfer.To = accountKeys[instruction.Accounts[1]]                  // 接收方账户
			transfer.Auth = accountKeys[instruction.Accounts[2]]                // 授权签名账户
			transfer.Amount = binary.LittleEndian.Uint64(instruction.Data[1:9]) // 转账金额
		} else if instruction.Data[0] == byte(token.InstructionTransferChecked) {
			// transferChecked指令长度一共10个字节
			if len(instruction.Data) < 10 {
				err = errors.New("data len not equal 10")
				return
			}
			if len(instruction.Accounts) < 4 {
				err = errors.New("account len too small")
				return
			}
			transfer.From = accountKeys[instruction.Accounts[0]] // 发送方账户
			// mint := accountKeys[instruction.Accounts[1]]		// 代币Mint账户
			transfer.To = accountKeys[instruction.Accounts[2]]                   // 接收方账户
			transfer.Auth = accountKeys[instruction.Accounts[3]]                 // 授权签名账户
			transfer.Amount = binary.LittleEndian.Uint64(instruction.Data[1:10]) // 转账金额
			// decimal := instruction.Data[10]
		} else {
			err = errors.New("not transfer Instruction")
			return
		}
		return transfer, nil
	}

	if accountKeys[instruction.ProgramIDIndex].String() != ProgramStrToken { // 检查指令是否是代币程序 //TokenProgramID
		err = errors.New("not token program")
		return
	}
	if len(instruction.Accounts) < 3 {
		err = errors.New("not enough accounts")
		return
	}
	if len(instruction.Data) < 1 {
		err = errors.New("data len to0 small")
		return
	}
	if instruction.Data[0] == byte(token.InstructionTransfer) { // 检查指令是否是转账指令
		if len(instruction.Data) != 9 {
			err = errors.New("data len not equal 9")
			return
		}
		if len(instruction.Accounts) < 3 {
			err = errors.New("account len too small")
			return
		}
		transfer.From = accountKeys[instruction.Accounts[0]]                // 发送方账户
		transfer.To = accountKeys[instruction.Accounts[1]]                  // 接收方账户
		transfer.Auth = accountKeys[instruction.Accounts[2]]                // 授权签名账户
		transfer.Amount = binary.LittleEndian.Uint64(instruction.Data[1:9]) // 转账金额
	} else if instruction.Data[0] == byte(token.InstructionTransferChecked) { // 检查指令是否是转账指令（Checked）
		if len(instruction.Data) != 10 {
			err = errors.New("data len not equal 10")
			return
		}
		if len(instruction.Accounts) < 4 {
			err = errors.New("account len too small")
			return
		}
		transfer.From = accountKeys[instruction.Accounts[0]] // 发送方账户
		// mint := accountKeys[instruction.Accounts[1]] // 代币Mint账户
		transfer.To = accountKeys[instruction.Accounts[2]]                   // 接收方账户
		transfer.Auth = accountKeys[instruction.Accounts[3]]                 // 授权签名账户
		transfer.Amount = binary.LittleEndian.Uint64(instruction.Data[1:10]) // 转账金额
		// decimal := instruction.Data[10]
	} else {
		err = errors.New("not transfer Instruction")
		return
	}

	return
}
