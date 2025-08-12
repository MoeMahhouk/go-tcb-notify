// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bindings

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// IFlashtestationRegistryRegisteredTEE is an auto generated low-level Go binding around an user-defined struct.
type IFlashtestationRegistryRegisteredTEE struct {
	IsValid                  bool
	RawQuote                 []byte
	ParsedReportBody         TD10ReportBody
	ExtendedRegistrationData []byte
}

// TD10ReportBody is an auto generated low-level Go binding around an user-defined struct.
type TD10ReportBody struct {
	TeeTcbSvn      [16]byte
	MrSeam         []byte
	MrsignerSeam   []byte
	SeamAttributes [8]byte
	TdAttributes   [8]byte
	XFAM           [8]byte
	MrTd           []byte
	MrConfigId     []byte
	MrOwner        []byte
	MrOwnerConfig  []byte
	RtMr0          []byte
	RtMr1          []byte
	RtMr2          []byte
	RtMr3          []byte
	ReportData     []byte
}

// FlashtestationRegistryMetaData contains all meta data concerning the FlashtestationRegistry contract.
var FlashtestationRegistryMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"MAX_BYTES_SIZE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"REGISTER_TYPEHASH\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"TD_REPORTDATA_LENGTH\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"UPGRADE_INTERFACE_VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"attestationContract\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIAttestation\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"computeStructHash\",\"inputs\":[{\"name\":\"rawQuote\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"extendedRegistrationData\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nonce\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"eip712Domain\",\"inputs\":[],\"outputs\":[{\"name\":\"fields\",\"type\":\"bytes1\",\"internalType\":\"bytes1\"},{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"chainId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"verifyingContract\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"salt\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"extensions\",\"type\":\"uint256[]\",\"internalType\":\"uint256[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRegistration\",\"inputs\":[{\"name\":\"teeAddress\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIFlashtestationRegistry.RegisteredTEE\",\"components\":[{\"name\":\"isValid\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"rawQuote\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"parsedReportBody\",\"type\":\"tuple\",\"internalType\":\"structTD10ReportBody\",\"components\":[{\"name\":\"teeTcbSvn\",\"type\":\"bytes16\",\"internalType\":\"bytes16\"},{\"name\":\"mrSeam\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"mrsignerSeam\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"seamAttributes\",\"type\":\"bytes8\",\"internalType\":\"bytes8\"},{\"name\":\"tdAttributes\",\"type\":\"bytes8\",\"internalType\":\"bytes8\"},{\"name\":\"xFAM\",\"type\":\"bytes8\",\"internalType\":\"bytes8\"},{\"name\":\"mrTd\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"mrConfigId\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"mrOwner\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"mrOwnerConfig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"rtMr0\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"rtMr1\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"rtMr2\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"rtMr3\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"reportData\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"extendedRegistrationData\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"hashTypedDataV4\",\"inputs\":[{\"name\":\"structHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"initialize\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_attestationContract\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"invalidateAttestation\",\"inputs\":[{\"name\":\"teeAddress\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"nonces\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"permitRegisterTEEService\",\"inputs\":[{\"name\":\"rawQuote\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"extendedRegistrationData\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nonce\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"signature\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"proxiableUUID\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"registerTEEService\",\"inputs\":[{\"name\":\"rawQuote\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"extendedRegistrationData\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"registeredTEEs\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"isValid\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"rawQuote\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"parsedReportBody\",\"type\":\"tuple\",\"internalType\":\"structTD10ReportBody\",\"components\":[{\"name\":\"teeTcbSvn\",\"type\":\"bytes16\",\"internalType\":\"bytes16\"},{\"name\":\"mrSeam\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"mrsignerSeam\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"seamAttributes\",\"type\":\"bytes8\",\"internalType\":\"bytes8\"},{\"name\":\"tdAttributes\",\"type\":\"bytes8\",\"internalType\":\"bytes8\"},{\"name\":\"xFAM\",\"type\":\"bytes8\",\"internalType\":\"bytes8\"},{\"name\":\"mrTd\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"mrConfigId\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"mrOwner\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"mrOwnerConfig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"rtMr0\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"rtMr1\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"rtMr2\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"rtMr3\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"reportData\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"extendedRegistrationData\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"upgradeToAndCall\",\"inputs\":[{\"name\":\"newImplementation\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"data\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"event\",\"name\":\"EIP712DomainChanged\",\"inputs\":[],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Initialized\",\"inputs\":[{\"name\":\"version\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TEEServiceInvalidated\",\"inputs\":[{\"name\":\"teeAddress\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TEEServiceRegistered\",\"inputs\":[{\"name\":\"teeAddress\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"rawQuote\",\"type\":\"bytes\",\"indexed\":false,\"internalType\":\"bytes\"},{\"name\":\"alreadyExists\",\"type\":\"bool\",\"indexed\":false,\"internalType\":\"bool\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Upgraded\",\"inputs\":[{\"name\":\"implementation\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"AddressEmptyCode\",\"inputs\":[{\"name\":\"target\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ByteSizeExceeded\",\"inputs\":[{\"name\":\"size\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"ECDSAInvalidSignature\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ECDSAInvalidSignatureLength\",\"inputs\":[{\"name\":\"length\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"ECDSAInvalidSignatureS\",\"inputs\":[{\"name\":\"s\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"ERC1967InvalidImplementation\",\"inputs\":[{\"name\":\"implementation\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ERC1967NonPayable\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"FailedCall\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidInitialization\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidNonce\",\"inputs\":[{\"name\":\"expected\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"provided\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"InvalidQuote\",\"inputs\":[{\"name\":\"output\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"type\":\"error\",\"name\":\"InvalidQuoteLength\",\"inputs\":[{\"name\":\"length\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"InvalidRegistrationDataHash\",\"inputs\":[{\"name\":\"expected\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"received\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"InvalidReportDataLength\",\"inputs\":[{\"name\":\"length\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"InvalidTEEType\",\"inputs\":[{\"name\":\"teeType\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}]},{\"type\":\"error\",\"name\":\"InvalidTEEVersion\",\"inputs\":[{\"name\":\"version\",\"type\":\"uint16\",\"internalType\":\"uint16\"}]},{\"type\":\"error\",\"name\":\"NotInitializing\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OwnableInvalidOwner\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"OwnableUnauthorizedAccount\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ReentrancyGuardReentrantCall\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"SenderMustMatchTEEAddress\",\"inputs\":[{\"name\":\"sender\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"teeAddress\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"TEEIsStillValid\",\"inputs\":[{\"name\":\"teeAddress\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"TEEServiceAlreadyInvalid\",\"inputs\":[{\"name\":\"teeAddress\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"TEEServiceAlreadyRegistered\",\"inputs\":[{\"name\":\"teeAddress\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"TEEServiceNotRegistered\",\"inputs\":[{\"name\":\"teeAddress\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"UUPSUnauthorizedCallContext\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"UUPSUnsupportedProxiableUUID\",\"inputs\":[{\"name\":\"slot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]}]",
}

// FlashtestationRegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use FlashtestationRegistryMetaData.ABI instead.
var FlashtestationRegistryABI = FlashtestationRegistryMetaData.ABI

// FlashtestationRegistry is an auto generated Go binding around an Ethereum contract.
type FlashtestationRegistry struct {
	FlashtestationRegistryCaller     // Read-only binding to the contract
	FlashtestationRegistryTransactor // Write-only binding to the contract
	FlashtestationRegistryFilterer   // Log filterer for contract events
}

// FlashtestationRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type FlashtestationRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FlashtestationRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type FlashtestationRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FlashtestationRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type FlashtestationRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FlashtestationRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type FlashtestationRegistrySession struct {
	Contract     *FlashtestationRegistry // Generic contract binding to set the session for
	CallOpts     bind.CallOpts           // Call options to use throughout this session
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// FlashtestationRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type FlashtestationRegistryCallerSession struct {
	Contract *FlashtestationRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                 // Call options to use throughout this session
}

// FlashtestationRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type FlashtestationRegistryTransactorSession struct {
	Contract     *FlashtestationRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                 // Transaction auth options to use throughout this session
}

// FlashtestationRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type FlashtestationRegistryRaw struct {
	Contract *FlashtestationRegistry // Generic contract binding to access the raw methods on
}

// FlashtestationRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type FlashtestationRegistryCallerRaw struct {
	Contract *FlashtestationRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// FlashtestationRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type FlashtestationRegistryTransactorRaw struct {
	Contract *FlashtestationRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewFlashtestationRegistry creates a new instance of FlashtestationRegistry, bound to a specific deployed contract.
func NewFlashtestationRegistry(address common.Address, backend bind.ContractBackend) (*FlashtestationRegistry, error) {
	contract, err := bindFlashtestationRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &FlashtestationRegistry{FlashtestationRegistryCaller: FlashtestationRegistryCaller{contract: contract}, FlashtestationRegistryTransactor: FlashtestationRegistryTransactor{contract: contract}, FlashtestationRegistryFilterer: FlashtestationRegistryFilterer{contract: contract}}, nil
}

// NewFlashtestationRegistryCaller creates a new read-only instance of FlashtestationRegistry, bound to a specific deployed contract.
func NewFlashtestationRegistryCaller(address common.Address, caller bind.ContractCaller) (*FlashtestationRegistryCaller, error) {
	contract, err := bindFlashtestationRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &FlashtestationRegistryCaller{contract: contract}, nil
}

// NewFlashtestationRegistryTransactor creates a new write-only instance of FlashtestationRegistry, bound to a specific deployed contract.
func NewFlashtestationRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*FlashtestationRegistryTransactor, error) {
	contract, err := bindFlashtestationRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &FlashtestationRegistryTransactor{contract: contract}, nil
}

// NewFlashtestationRegistryFilterer creates a new log filterer instance of FlashtestationRegistry, bound to a specific deployed contract.
func NewFlashtestationRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*FlashtestationRegistryFilterer, error) {
	contract, err := bindFlashtestationRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &FlashtestationRegistryFilterer{contract: contract}, nil
}

// bindFlashtestationRegistry binds a generic wrapper to an already deployed contract.
func bindFlashtestationRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := FlashtestationRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_FlashtestationRegistry *FlashtestationRegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _FlashtestationRegistry.Contract.FlashtestationRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_FlashtestationRegistry *FlashtestationRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _FlashtestationRegistry.Contract.FlashtestationRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_FlashtestationRegistry *FlashtestationRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _FlashtestationRegistry.Contract.FlashtestationRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_FlashtestationRegistry *FlashtestationRegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _FlashtestationRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_FlashtestationRegistry *FlashtestationRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _FlashtestationRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_FlashtestationRegistry *FlashtestationRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _FlashtestationRegistry.Contract.contract.Transact(opts, method, params...)
}

// MAXBYTESSIZE is a free data retrieval call binding the contract method 0xaaae748e.
//
// Solidity: function MAX_BYTES_SIZE() view returns(uint256)
func (_FlashtestationRegistry *FlashtestationRegistryCaller) MAXBYTESSIZE(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _FlashtestationRegistry.contract.Call(opts, &out, "MAX_BYTES_SIZE")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MAXBYTESSIZE is a free data retrieval call binding the contract method 0xaaae748e.
//
// Solidity: function MAX_BYTES_SIZE() view returns(uint256)
func (_FlashtestationRegistry *FlashtestationRegistrySession) MAXBYTESSIZE() (*big.Int, error) {
	return _FlashtestationRegistry.Contract.MAXBYTESSIZE(&_FlashtestationRegistry.CallOpts)
}

// MAXBYTESSIZE is a free data retrieval call binding the contract method 0xaaae748e.
//
// Solidity: function MAX_BYTES_SIZE() view returns(uint256)
func (_FlashtestationRegistry *FlashtestationRegistryCallerSession) MAXBYTESSIZE() (*big.Int, error) {
	return _FlashtestationRegistry.Contract.MAXBYTESSIZE(&_FlashtestationRegistry.CallOpts)
}

// REGISTERTYPEHASH is a free data retrieval call binding the contract method 0x6a5306a3.
//
// Solidity: function REGISTER_TYPEHASH() view returns(bytes32)
func (_FlashtestationRegistry *FlashtestationRegistryCaller) REGISTERTYPEHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _FlashtestationRegistry.contract.Call(opts, &out, "REGISTER_TYPEHASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// REGISTERTYPEHASH is a free data retrieval call binding the contract method 0x6a5306a3.
//
// Solidity: function REGISTER_TYPEHASH() view returns(bytes32)
func (_FlashtestationRegistry *FlashtestationRegistrySession) REGISTERTYPEHASH() ([32]byte, error) {
	return _FlashtestationRegistry.Contract.REGISTERTYPEHASH(&_FlashtestationRegistry.CallOpts)
}

// REGISTERTYPEHASH is a free data retrieval call binding the contract method 0x6a5306a3.
//
// Solidity: function REGISTER_TYPEHASH() view returns(bytes32)
func (_FlashtestationRegistry *FlashtestationRegistryCallerSession) REGISTERTYPEHASH() ([32]byte, error) {
	return _FlashtestationRegistry.Contract.REGISTERTYPEHASH(&_FlashtestationRegistry.CallOpts)
}

// TDREPORTDATALENGTH is a free data retrieval call binding the contract method 0xe4168952.
//
// Solidity: function TD_REPORTDATA_LENGTH() view returns(uint256)
func (_FlashtestationRegistry *FlashtestationRegistryCaller) TDREPORTDATALENGTH(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _FlashtestationRegistry.contract.Call(opts, &out, "TD_REPORTDATA_LENGTH")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TDREPORTDATALENGTH is a free data retrieval call binding the contract method 0xe4168952.
//
// Solidity: function TD_REPORTDATA_LENGTH() view returns(uint256)
func (_FlashtestationRegistry *FlashtestationRegistrySession) TDREPORTDATALENGTH() (*big.Int, error) {
	return _FlashtestationRegistry.Contract.TDREPORTDATALENGTH(&_FlashtestationRegistry.CallOpts)
}

// TDREPORTDATALENGTH is a free data retrieval call binding the contract method 0xe4168952.
//
// Solidity: function TD_REPORTDATA_LENGTH() view returns(uint256)
func (_FlashtestationRegistry *FlashtestationRegistryCallerSession) TDREPORTDATALENGTH() (*big.Int, error) {
	return _FlashtestationRegistry.Contract.TDREPORTDATALENGTH(&_FlashtestationRegistry.CallOpts)
}

// UPGRADEINTERFACEVERSION is a free data retrieval call binding the contract method 0xad3cb1cc.
//
// Solidity: function UPGRADE_INTERFACE_VERSION() view returns(string)
func (_FlashtestationRegistry *FlashtestationRegistryCaller) UPGRADEINTERFACEVERSION(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _FlashtestationRegistry.contract.Call(opts, &out, "UPGRADE_INTERFACE_VERSION")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// UPGRADEINTERFACEVERSION is a free data retrieval call binding the contract method 0xad3cb1cc.
//
// Solidity: function UPGRADE_INTERFACE_VERSION() view returns(string)
func (_FlashtestationRegistry *FlashtestationRegistrySession) UPGRADEINTERFACEVERSION() (string, error) {
	return _FlashtestationRegistry.Contract.UPGRADEINTERFACEVERSION(&_FlashtestationRegistry.CallOpts)
}

// UPGRADEINTERFACEVERSION is a free data retrieval call binding the contract method 0xad3cb1cc.
//
// Solidity: function UPGRADE_INTERFACE_VERSION() view returns(string)
func (_FlashtestationRegistry *FlashtestationRegistryCallerSession) UPGRADEINTERFACEVERSION() (string, error) {
	return _FlashtestationRegistry.Contract.UPGRADEINTERFACEVERSION(&_FlashtestationRegistry.CallOpts)
}

// AttestationContract is a free data retrieval call binding the contract method 0x87be6d4e.
//
// Solidity: function attestationContract() view returns(address)
func (_FlashtestationRegistry *FlashtestationRegistryCaller) AttestationContract(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _FlashtestationRegistry.contract.Call(opts, &out, "attestationContract")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// AttestationContract is a free data retrieval call binding the contract method 0x87be6d4e.
//
// Solidity: function attestationContract() view returns(address)
func (_FlashtestationRegistry *FlashtestationRegistrySession) AttestationContract() (common.Address, error) {
	return _FlashtestationRegistry.Contract.AttestationContract(&_FlashtestationRegistry.CallOpts)
}

// AttestationContract is a free data retrieval call binding the contract method 0x87be6d4e.
//
// Solidity: function attestationContract() view returns(address)
func (_FlashtestationRegistry *FlashtestationRegistryCallerSession) AttestationContract() (common.Address, error) {
	return _FlashtestationRegistry.Contract.AttestationContract(&_FlashtestationRegistry.CallOpts)
}

// ComputeStructHash is a free data retrieval call binding the contract method 0x22a43e25.
//
// Solidity: function computeStructHash(bytes rawQuote, bytes extendedRegistrationData, uint256 nonce) pure returns(bytes32)
func (_FlashtestationRegistry *FlashtestationRegistryCaller) ComputeStructHash(opts *bind.CallOpts, rawQuote []byte, extendedRegistrationData []byte, nonce *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _FlashtestationRegistry.contract.Call(opts, &out, "computeStructHash", rawQuote, extendedRegistrationData, nonce)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ComputeStructHash is a free data retrieval call binding the contract method 0x22a43e25.
//
// Solidity: function computeStructHash(bytes rawQuote, bytes extendedRegistrationData, uint256 nonce) pure returns(bytes32)
func (_FlashtestationRegistry *FlashtestationRegistrySession) ComputeStructHash(rawQuote []byte, extendedRegistrationData []byte, nonce *big.Int) ([32]byte, error) {
	return _FlashtestationRegistry.Contract.ComputeStructHash(&_FlashtestationRegistry.CallOpts, rawQuote, extendedRegistrationData, nonce)
}

// ComputeStructHash is a free data retrieval call binding the contract method 0x22a43e25.
//
// Solidity: function computeStructHash(bytes rawQuote, bytes extendedRegistrationData, uint256 nonce) pure returns(bytes32)
func (_FlashtestationRegistry *FlashtestationRegistryCallerSession) ComputeStructHash(rawQuote []byte, extendedRegistrationData []byte, nonce *big.Int) ([32]byte, error) {
	return _FlashtestationRegistry.Contract.ComputeStructHash(&_FlashtestationRegistry.CallOpts, rawQuote, extendedRegistrationData, nonce)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_FlashtestationRegistry *FlashtestationRegistryCaller) Eip712Domain(opts *bind.CallOpts) (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	var out []interface{}
	err := _FlashtestationRegistry.contract.Call(opts, &out, "eip712Domain")

	outstruct := new(struct {
		Fields            [1]byte
		Name              string
		Version           string
		ChainId           *big.Int
		VerifyingContract common.Address
		Salt              [32]byte
		Extensions        []*big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Fields = *abi.ConvertType(out[0], new([1]byte)).(*[1]byte)
	outstruct.Name = *abi.ConvertType(out[1], new(string)).(*string)
	outstruct.Version = *abi.ConvertType(out[2], new(string)).(*string)
	outstruct.ChainId = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.VerifyingContract = *abi.ConvertType(out[4], new(common.Address)).(*common.Address)
	outstruct.Salt = *abi.ConvertType(out[5], new([32]byte)).(*[32]byte)
	outstruct.Extensions = *abi.ConvertType(out[6], new([]*big.Int)).(*[]*big.Int)

	return *outstruct, err

}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_FlashtestationRegistry *FlashtestationRegistrySession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _FlashtestationRegistry.Contract.Eip712Domain(&_FlashtestationRegistry.CallOpts)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_FlashtestationRegistry *FlashtestationRegistryCallerSession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _FlashtestationRegistry.Contract.Eip712Domain(&_FlashtestationRegistry.CallOpts)
}

// GetRegistration is a free data retrieval call binding the contract method 0x72731062.
//
// Solidity: function getRegistration(address teeAddress) view returns(bool, (bool,bytes,(bytes16,bytes,bytes,bytes8,bytes8,bytes8,bytes,bytes,bytes,bytes,bytes,bytes,bytes,bytes,bytes),bytes))
func (_FlashtestationRegistry *FlashtestationRegistryCaller) GetRegistration(opts *bind.CallOpts, teeAddress common.Address) (bool, IFlashtestationRegistryRegisteredTEE, error) {
	var out []interface{}
	err := _FlashtestationRegistry.contract.Call(opts, &out, "getRegistration", teeAddress)

	if err != nil {
		return *new(bool), *new(IFlashtestationRegistryRegisteredTEE), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)
	out1 := *abi.ConvertType(out[1], new(IFlashtestationRegistryRegisteredTEE)).(*IFlashtestationRegistryRegisteredTEE)

	return out0, out1, err

}

// GetRegistration is a free data retrieval call binding the contract method 0x72731062.
//
// Solidity: function getRegistration(address teeAddress) view returns(bool, (bool,bytes,(bytes16,bytes,bytes,bytes8,bytes8,bytes8,bytes,bytes,bytes,bytes,bytes,bytes,bytes,bytes,bytes),bytes))
func (_FlashtestationRegistry *FlashtestationRegistrySession) GetRegistration(teeAddress common.Address) (bool, IFlashtestationRegistryRegisteredTEE, error) {
	return _FlashtestationRegistry.Contract.GetRegistration(&_FlashtestationRegistry.CallOpts, teeAddress)
}

// GetRegistration is a free data retrieval call binding the contract method 0x72731062.
//
// Solidity: function getRegistration(address teeAddress) view returns(bool, (bool,bytes,(bytes16,bytes,bytes,bytes8,bytes8,bytes8,bytes,bytes,bytes,bytes,bytes,bytes,bytes,bytes,bytes),bytes))
func (_FlashtestationRegistry *FlashtestationRegistryCallerSession) GetRegistration(teeAddress common.Address) (bool, IFlashtestationRegistryRegisteredTEE, error) {
	return _FlashtestationRegistry.Contract.GetRegistration(&_FlashtestationRegistry.CallOpts, teeAddress)
}

// HashTypedDataV4 is a free data retrieval call binding the contract method 0x4980f288.
//
// Solidity: function hashTypedDataV4(bytes32 structHash) view returns(bytes32)
func (_FlashtestationRegistry *FlashtestationRegistryCaller) HashTypedDataV4(opts *bind.CallOpts, structHash [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _FlashtestationRegistry.contract.Call(opts, &out, "hashTypedDataV4", structHash)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// HashTypedDataV4 is a free data retrieval call binding the contract method 0x4980f288.
//
// Solidity: function hashTypedDataV4(bytes32 structHash) view returns(bytes32)
func (_FlashtestationRegistry *FlashtestationRegistrySession) HashTypedDataV4(structHash [32]byte) ([32]byte, error) {
	return _FlashtestationRegistry.Contract.HashTypedDataV4(&_FlashtestationRegistry.CallOpts, structHash)
}

// HashTypedDataV4 is a free data retrieval call binding the contract method 0x4980f288.
//
// Solidity: function hashTypedDataV4(bytes32 structHash) view returns(bytes32)
func (_FlashtestationRegistry *FlashtestationRegistryCallerSession) HashTypedDataV4(structHash [32]byte) ([32]byte, error) {
	return _FlashtestationRegistry.Contract.HashTypedDataV4(&_FlashtestationRegistry.CallOpts, structHash)
}

// Nonces is a free data retrieval call binding the contract method 0x7ecebe00.
//
// Solidity: function nonces(address ) view returns(uint256)
func (_FlashtestationRegistry *FlashtestationRegistryCaller) Nonces(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _FlashtestationRegistry.contract.Call(opts, &out, "nonces", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Nonces is a free data retrieval call binding the contract method 0x7ecebe00.
//
// Solidity: function nonces(address ) view returns(uint256)
func (_FlashtestationRegistry *FlashtestationRegistrySession) Nonces(arg0 common.Address) (*big.Int, error) {
	return _FlashtestationRegistry.Contract.Nonces(&_FlashtestationRegistry.CallOpts, arg0)
}

// Nonces is a free data retrieval call binding the contract method 0x7ecebe00.
//
// Solidity: function nonces(address ) view returns(uint256)
func (_FlashtestationRegistry *FlashtestationRegistryCallerSession) Nonces(arg0 common.Address) (*big.Int, error) {
	return _FlashtestationRegistry.Contract.Nonces(&_FlashtestationRegistry.CallOpts, arg0)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_FlashtestationRegistry *FlashtestationRegistryCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _FlashtestationRegistry.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_FlashtestationRegistry *FlashtestationRegistrySession) Owner() (common.Address, error) {
	return _FlashtestationRegistry.Contract.Owner(&_FlashtestationRegistry.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_FlashtestationRegistry *FlashtestationRegistryCallerSession) Owner() (common.Address, error) {
	return _FlashtestationRegistry.Contract.Owner(&_FlashtestationRegistry.CallOpts)
}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_FlashtestationRegistry *FlashtestationRegistryCaller) ProxiableUUID(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _FlashtestationRegistry.contract.Call(opts, &out, "proxiableUUID")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_FlashtestationRegistry *FlashtestationRegistrySession) ProxiableUUID() ([32]byte, error) {
	return _FlashtestationRegistry.Contract.ProxiableUUID(&_FlashtestationRegistry.CallOpts)
}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_FlashtestationRegistry *FlashtestationRegistryCallerSession) ProxiableUUID() ([32]byte, error) {
	return _FlashtestationRegistry.Contract.ProxiableUUID(&_FlashtestationRegistry.CallOpts)
}

// RegisteredTEEs is a free data retrieval call binding the contract method 0xf745cb30.
//
// Solidity: function registeredTEEs(address ) view returns(bool isValid, bytes rawQuote, (bytes16,bytes,bytes,bytes8,bytes8,bytes8,bytes,bytes,bytes,bytes,bytes,bytes,bytes,bytes,bytes) parsedReportBody, bytes extendedRegistrationData)
func (_FlashtestationRegistry *FlashtestationRegistryCaller) RegisteredTEEs(opts *bind.CallOpts, arg0 common.Address) (struct {
	IsValid                  bool
	RawQuote                 []byte
	ParsedReportBody         TD10ReportBody
	ExtendedRegistrationData []byte
}, error) {
	var out []interface{}
	err := _FlashtestationRegistry.contract.Call(opts, &out, "registeredTEEs", arg0)

	outstruct := new(struct {
		IsValid                  bool
		RawQuote                 []byte
		ParsedReportBody         TD10ReportBody
		ExtendedRegistrationData []byte
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.IsValid = *abi.ConvertType(out[0], new(bool)).(*bool)
	outstruct.RawQuote = *abi.ConvertType(out[1], new([]byte)).(*[]byte)
	outstruct.ParsedReportBody = *abi.ConvertType(out[2], new(TD10ReportBody)).(*TD10ReportBody)
	outstruct.ExtendedRegistrationData = *abi.ConvertType(out[3], new([]byte)).(*[]byte)

	return *outstruct, err

}

// RegisteredTEEs is a free data retrieval call binding the contract method 0xf745cb30.
//
// Solidity: function registeredTEEs(address ) view returns(bool isValid, bytes rawQuote, (bytes16,bytes,bytes,bytes8,bytes8,bytes8,bytes,bytes,bytes,bytes,bytes,bytes,bytes,bytes,bytes) parsedReportBody, bytes extendedRegistrationData)
func (_FlashtestationRegistry *FlashtestationRegistrySession) RegisteredTEEs(arg0 common.Address) (struct {
	IsValid                  bool
	RawQuote                 []byte
	ParsedReportBody         TD10ReportBody
	ExtendedRegistrationData []byte
}, error) {
	return _FlashtestationRegistry.Contract.RegisteredTEEs(&_FlashtestationRegistry.CallOpts, arg0)
}

// RegisteredTEEs is a free data retrieval call binding the contract method 0xf745cb30.
//
// Solidity: function registeredTEEs(address ) view returns(bool isValid, bytes rawQuote, (bytes16,bytes,bytes,bytes8,bytes8,bytes8,bytes,bytes,bytes,bytes,bytes,bytes,bytes,bytes,bytes) parsedReportBody, bytes extendedRegistrationData)
func (_FlashtestationRegistry *FlashtestationRegistryCallerSession) RegisteredTEEs(arg0 common.Address) (struct {
	IsValid                  bool
	RawQuote                 []byte
	ParsedReportBody         TD10ReportBody
	ExtendedRegistrationData []byte
}, error) {
	return _FlashtestationRegistry.Contract.RegisteredTEEs(&_FlashtestationRegistry.CallOpts, arg0)
}

// Initialize is a paid mutator transaction binding the contract method 0x485cc955.
//
// Solidity: function initialize(address owner, address _attestationContract) returns()
func (_FlashtestationRegistry *FlashtestationRegistryTransactor) Initialize(opts *bind.TransactOpts, owner common.Address, _attestationContract common.Address) (*types.Transaction, error) {
	return _FlashtestationRegistry.contract.Transact(opts, "initialize", owner, _attestationContract)
}

// Initialize is a paid mutator transaction binding the contract method 0x485cc955.
//
// Solidity: function initialize(address owner, address _attestationContract) returns()
func (_FlashtestationRegistry *FlashtestationRegistrySession) Initialize(owner common.Address, _attestationContract common.Address) (*types.Transaction, error) {
	return _FlashtestationRegistry.Contract.Initialize(&_FlashtestationRegistry.TransactOpts, owner, _attestationContract)
}

// Initialize is a paid mutator transaction binding the contract method 0x485cc955.
//
// Solidity: function initialize(address owner, address _attestationContract) returns()
func (_FlashtestationRegistry *FlashtestationRegistryTransactorSession) Initialize(owner common.Address, _attestationContract common.Address) (*types.Transaction, error) {
	return _FlashtestationRegistry.Contract.Initialize(&_FlashtestationRegistry.TransactOpts, owner, _attestationContract)
}

// InvalidateAttestation is a paid mutator transaction binding the contract method 0xf9b68b31.
//
// Solidity: function invalidateAttestation(address teeAddress) returns()
func (_FlashtestationRegistry *FlashtestationRegistryTransactor) InvalidateAttestation(opts *bind.TransactOpts, teeAddress common.Address) (*types.Transaction, error) {
	return _FlashtestationRegistry.contract.Transact(opts, "invalidateAttestation", teeAddress)
}

// InvalidateAttestation is a paid mutator transaction binding the contract method 0xf9b68b31.
//
// Solidity: function invalidateAttestation(address teeAddress) returns()
func (_FlashtestationRegistry *FlashtestationRegistrySession) InvalidateAttestation(teeAddress common.Address) (*types.Transaction, error) {
	return _FlashtestationRegistry.Contract.InvalidateAttestation(&_FlashtestationRegistry.TransactOpts, teeAddress)
}

// InvalidateAttestation is a paid mutator transaction binding the contract method 0xf9b68b31.
//
// Solidity: function invalidateAttestation(address teeAddress) returns()
func (_FlashtestationRegistry *FlashtestationRegistryTransactorSession) InvalidateAttestation(teeAddress common.Address) (*types.Transaction, error) {
	return _FlashtestationRegistry.Contract.InvalidateAttestation(&_FlashtestationRegistry.TransactOpts, teeAddress)
}

// PermitRegisterTEEService is a paid mutator transaction binding the contract method 0x4111c12b.
//
// Solidity: function permitRegisterTEEService(bytes rawQuote, bytes extendedRegistrationData, uint256 nonce, bytes signature) returns()
func (_FlashtestationRegistry *FlashtestationRegistryTransactor) PermitRegisterTEEService(opts *bind.TransactOpts, rawQuote []byte, extendedRegistrationData []byte, nonce *big.Int, signature []byte) (*types.Transaction, error) {
	return _FlashtestationRegistry.contract.Transact(opts, "permitRegisterTEEService", rawQuote, extendedRegistrationData, nonce, signature)
}

// PermitRegisterTEEService is a paid mutator transaction binding the contract method 0x4111c12b.
//
// Solidity: function permitRegisterTEEService(bytes rawQuote, bytes extendedRegistrationData, uint256 nonce, bytes signature) returns()
func (_FlashtestationRegistry *FlashtestationRegistrySession) PermitRegisterTEEService(rawQuote []byte, extendedRegistrationData []byte, nonce *big.Int, signature []byte) (*types.Transaction, error) {
	return _FlashtestationRegistry.Contract.PermitRegisterTEEService(&_FlashtestationRegistry.TransactOpts, rawQuote, extendedRegistrationData, nonce, signature)
}

// PermitRegisterTEEService is a paid mutator transaction binding the contract method 0x4111c12b.
//
// Solidity: function permitRegisterTEEService(bytes rawQuote, bytes extendedRegistrationData, uint256 nonce, bytes signature) returns()
func (_FlashtestationRegistry *FlashtestationRegistryTransactorSession) PermitRegisterTEEService(rawQuote []byte, extendedRegistrationData []byte, nonce *big.Int, signature []byte) (*types.Transaction, error) {
	return _FlashtestationRegistry.Contract.PermitRegisterTEEService(&_FlashtestationRegistry.TransactOpts, rawQuote, extendedRegistrationData, nonce, signature)
}

// RegisterTEEService is a paid mutator transaction binding the contract method 0x22ba2bbf.
//
// Solidity: function registerTEEService(bytes rawQuote, bytes extendedRegistrationData) returns()
func (_FlashtestationRegistry *FlashtestationRegistryTransactor) RegisterTEEService(opts *bind.TransactOpts, rawQuote []byte, extendedRegistrationData []byte) (*types.Transaction, error) {
	return _FlashtestationRegistry.contract.Transact(opts, "registerTEEService", rawQuote, extendedRegistrationData)
}

// RegisterTEEService is a paid mutator transaction binding the contract method 0x22ba2bbf.
//
// Solidity: function registerTEEService(bytes rawQuote, bytes extendedRegistrationData) returns()
func (_FlashtestationRegistry *FlashtestationRegistrySession) RegisterTEEService(rawQuote []byte, extendedRegistrationData []byte) (*types.Transaction, error) {
	return _FlashtestationRegistry.Contract.RegisterTEEService(&_FlashtestationRegistry.TransactOpts, rawQuote, extendedRegistrationData)
}

// RegisterTEEService is a paid mutator transaction binding the contract method 0x22ba2bbf.
//
// Solidity: function registerTEEService(bytes rawQuote, bytes extendedRegistrationData) returns()
func (_FlashtestationRegistry *FlashtestationRegistryTransactorSession) RegisterTEEService(rawQuote []byte, extendedRegistrationData []byte) (*types.Transaction, error) {
	return _FlashtestationRegistry.Contract.RegisterTEEService(&_FlashtestationRegistry.TransactOpts, rawQuote, extendedRegistrationData)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_FlashtestationRegistry *FlashtestationRegistryTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _FlashtestationRegistry.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_FlashtestationRegistry *FlashtestationRegistrySession) RenounceOwnership() (*types.Transaction, error) {
	return _FlashtestationRegistry.Contract.RenounceOwnership(&_FlashtestationRegistry.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_FlashtestationRegistry *FlashtestationRegistryTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _FlashtestationRegistry.Contract.RenounceOwnership(&_FlashtestationRegistry.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_FlashtestationRegistry *FlashtestationRegistryTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _FlashtestationRegistry.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_FlashtestationRegistry *FlashtestationRegistrySession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _FlashtestationRegistry.Contract.TransferOwnership(&_FlashtestationRegistry.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_FlashtestationRegistry *FlashtestationRegistryTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _FlashtestationRegistry.Contract.TransferOwnership(&_FlashtestationRegistry.TransactOpts, newOwner)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_FlashtestationRegistry *FlashtestationRegistryTransactor) UpgradeToAndCall(opts *bind.TransactOpts, newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _FlashtestationRegistry.contract.Transact(opts, "upgradeToAndCall", newImplementation, data)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_FlashtestationRegistry *FlashtestationRegistrySession) UpgradeToAndCall(newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _FlashtestationRegistry.Contract.UpgradeToAndCall(&_FlashtestationRegistry.TransactOpts, newImplementation, data)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_FlashtestationRegistry *FlashtestationRegistryTransactorSession) UpgradeToAndCall(newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _FlashtestationRegistry.Contract.UpgradeToAndCall(&_FlashtestationRegistry.TransactOpts, newImplementation, data)
}

// FlashtestationRegistryEIP712DomainChangedIterator is returned from FilterEIP712DomainChanged and is used to iterate over the raw logs and unpacked data for EIP712DomainChanged events raised by the FlashtestationRegistry contract.
type FlashtestationRegistryEIP712DomainChangedIterator struct {
	Event *FlashtestationRegistryEIP712DomainChanged // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *FlashtestationRegistryEIP712DomainChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FlashtestationRegistryEIP712DomainChanged)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(FlashtestationRegistryEIP712DomainChanged)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *FlashtestationRegistryEIP712DomainChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FlashtestationRegistryEIP712DomainChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FlashtestationRegistryEIP712DomainChanged represents a EIP712DomainChanged event raised by the FlashtestationRegistry contract.
type FlashtestationRegistryEIP712DomainChanged struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterEIP712DomainChanged is a free log retrieval operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_FlashtestationRegistry *FlashtestationRegistryFilterer) FilterEIP712DomainChanged(opts *bind.FilterOpts) (*FlashtestationRegistryEIP712DomainChangedIterator, error) {

	logs, sub, err := _FlashtestationRegistry.contract.FilterLogs(opts, "EIP712DomainChanged")
	if err != nil {
		return nil, err
	}
	return &FlashtestationRegistryEIP712DomainChangedIterator{contract: _FlashtestationRegistry.contract, event: "EIP712DomainChanged", logs: logs, sub: sub}, nil
}

// WatchEIP712DomainChanged is a free log subscription operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_FlashtestationRegistry *FlashtestationRegistryFilterer) WatchEIP712DomainChanged(opts *bind.WatchOpts, sink chan<- *FlashtestationRegistryEIP712DomainChanged) (event.Subscription, error) {

	logs, sub, err := _FlashtestationRegistry.contract.WatchLogs(opts, "EIP712DomainChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FlashtestationRegistryEIP712DomainChanged)
				if err := _FlashtestationRegistry.contract.UnpackLog(event, "EIP712DomainChanged", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseEIP712DomainChanged is a log parse operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_FlashtestationRegistry *FlashtestationRegistryFilterer) ParseEIP712DomainChanged(log types.Log) (*FlashtestationRegistryEIP712DomainChanged, error) {
	event := new(FlashtestationRegistryEIP712DomainChanged)
	if err := _FlashtestationRegistry.contract.UnpackLog(event, "EIP712DomainChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// FlashtestationRegistryInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the FlashtestationRegistry contract.
type FlashtestationRegistryInitializedIterator struct {
	Event *FlashtestationRegistryInitialized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *FlashtestationRegistryInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FlashtestationRegistryInitialized)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(FlashtestationRegistryInitialized)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *FlashtestationRegistryInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FlashtestationRegistryInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FlashtestationRegistryInitialized represents a Initialized event raised by the FlashtestationRegistry contract.
type FlashtestationRegistryInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_FlashtestationRegistry *FlashtestationRegistryFilterer) FilterInitialized(opts *bind.FilterOpts) (*FlashtestationRegistryInitializedIterator, error) {

	logs, sub, err := _FlashtestationRegistry.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &FlashtestationRegistryInitializedIterator{contract: _FlashtestationRegistry.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_FlashtestationRegistry *FlashtestationRegistryFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *FlashtestationRegistryInitialized) (event.Subscription, error) {

	logs, sub, err := _FlashtestationRegistry.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FlashtestationRegistryInitialized)
				if err := _FlashtestationRegistry.contract.UnpackLog(event, "Initialized", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseInitialized is a log parse operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_FlashtestationRegistry *FlashtestationRegistryFilterer) ParseInitialized(log types.Log) (*FlashtestationRegistryInitialized, error) {
	event := new(FlashtestationRegistryInitialized)
	if err := _FlashtestationRegistry.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// FlashtestationRegistryOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the FlashtestationRegistry contract.
type FlashtestationRegistryOwnershipTransferredIterator struct {
	Event *FlashtestationRegistryOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *FlashtestationRegistryOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FlashtestationRegistryOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(FlashtestationRegistryOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *FlashtestationRegistryOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FlashtestationRegistryOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FlashtestationRegistryOwnershipTransferred represents a OwnershipTransferred event raised by the FlashtestationRegistry contract.
type FlashtestationRegistryOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_FlashtestationRegistry *FlashtestationRegistryFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*FlashtestationRegistryOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _FlashtestationRegistry.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &FlashtestationRegistryOwnershipTransferredIterator{contract: _FlashtestationRegistry.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_FlashtestationRegistry *FlashtestationRegistryFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *FlashtestationRegistryOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _FlashtestationRegistry.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FlashtestationRegistryOwnershipTransferred)
				if err := _FlashtestationRegistry.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_FlashtestationRegistry *FlashtestationRegistryFilterer) ParseOwnershipTransferred(log types.Log) (*FlashtestationRegistryOwnershipTransferred, error) {
	event := new(FlashtestationRegistryOwnershipTransferred)
	if err := _FlashtestationRegistry.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// FlashtestationRegistryTEEServiceInvalidatedIterator is returned from FilterTEEServiceInvalidated and is used to iterate over the raw logs and unpacked data for TEEServiceInvalidated events raised by the FlashtestationRegistry contract.
type FlashtestationRegistryTEEServiceInvalidatedIterator struct {
	Event *FlashtestationRegistryTEEServiceInvalidated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *FlashtestationRegistryTEEServiceInvalidatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FlashtestationRegistryTEEServiceInvalidated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(FlashtestationRegistryTEEServiceInvalidated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *FlashtestationRegistryTEEServiceInvalidatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FlashtestationRegistryTEEServiceInvalidatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FlashtestationRegistryTEEServiceInvalidated represents a TEEServiceInvalidated event raised by the FlashtestationRegistry contract.
type FlashtestationRegistryTEEServiceInvalidated struct {
	TeeAddress common.Address
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterTEEServiceInvalidated is a free log retrieval operation binding the contract event 0x5bb0bbb0993a623e10dd3579bf5b9403deba943e0bfe950b740d60209c9135ef.
//
// Solidity: event TEEServiceInvalidated(address teeAddress)
func (_FlashtestationRegistry *FlashtestationRegistryFilterer) FilterTEEServiceInvalidated(opts *bind.FilterOpts) (*FlashtestationRegistryTEEServiceInvalidatedIterator, error) {

	logs, sub, err := _FlashtestationRegistry.contract.FilterLogs(opts, "TEEServiceInvalidated")
	if err != nil {
		return nil, err
	}
	return &FlashtestationRegistryTEEServiceInvalidatedIterator{contract: _FlashtestationRegistry.contract, event: "TEEServiceInvalidated", logs: logs, sub: sub}, nil
}

// WatchTEEServiceInvalidated is a free log subscription operation binding the contract event 0x5bb0bbb0993a623e10dd3579bf5b9403deba943e0bfe950b740d60209c9135ef.
//
// Solidity: event TEEServiceInvalidated(address teeAddress)
func (_FlashtestationRegistry *FlashtestationRegistryFilterer) WatchTEEServiceInvalidated(opts *bind.WatchOpts, sink chan<- *FlashtestationRegistryTEEServiceInvalidated) (event.Subscription, error) {

	logs, sub, err := _FlashtestationRegistry.contract.WatchLogs(opts, "TEEServiceInvalidated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FlashtestationRegistryTEEServiceInvalidated)
				if err := _FlashtestationRegistry.contract.UnpackLog(event, "TEEServiceInvalidated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseTEEServiceInvalidated is a log parse operation binding the contract event 0x5bb0bbb0993a623e10dd3579bf5b9403deba943e0bfe950b740d60209c9135ef.
//
// Solidity: event TEEServiceInvalidated(address teeAddress)
func (_FlashtestationRegistry *FlashtestationRegistryFilterer) ParseTEEServiceInvalidated(log types.Log) (*FlashtestationRegistryTEEServiceInvalidated, error) {
	event := new(FlashtestationRegistryTEEServiceInvalidated)
	if err := _FlashtestationRegistry.contract.UnpackLog(event, "TEEServiceInvalidated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// FlashtestationRegistryTEEServiceRegisteredIterator is returned from FilterTEEServiceRegistered and is used to iterate over the raw logs and unpacked data for TEEServiceRegistered events raised by the FlashtestationRegistry contract.
type FlashtestationRegistryTEEServiceRegisteredIterator struct {
	Event *FlashtestationRegistryTEEServiceRegistered // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *FlashtestationRegistryTEEServiceRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FlashtestationRegistryTEEServiceRegistered)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(FlashtestationRegistryTEEServiceRegistered)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *FlashtestationRegistryTEEServiceRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FlashtestationRegistryTEEServiceRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FlashtestationRegistryTEEServiceRegistered represents a TEEServiceRegistered event raised by the FlashtestationRegistry contract.
type FlashtestationRegistryTEEServiceRegistered struct {
	TeeAddress    common.Address
	RawQuote      []byte
	AlreadyExists bool
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterTEEServiceRegistered is a free log retrieval operation binding the contract event 0x206fdb1a74851a8542447b8b6704db24a36b906a7297cc23c2b984dc357b9978.
//
// Solidity: event TEEServiceRegistered(address teeAddress, bytes rawQuote, bool alreadyExists)
func (_FlashtestationRegistry *FlashtestationRegistryFilterer) FilterTEEServiceRegistered(opts *bind.FilterOpts) (*FlashtestationRegistryTEEServiceRegisteredIterator, error) {

	logs, sub, err := _FlashtestationRegistry.contract.FilterLogs(opts, "TEEServiceRegistered")
	if err != nil {
		return nil, err
	}
	return &FlashtestationRegistryTEEServiceRegisteredIterator{contract: _FlashtestationRegistry.contract, event: "TEEServiceRegistered", logs: logs, sub: sub}, nil
}

// WatchTEEServiceRegistered is a free log subscription operation binding the contract event 0x206fdb1a74851a8542447b8b6704db24a36b906a7297cc23c2b984dc357b9978.
//
// Solidity: event TEEServiceRegistered(address teeAddress, bytes rawQuote, bool alreadyExists)
func (_FlashtestationRegistry *FlashtestationRegistryFilterer) WatchTEEServiceRegistered(opts *bind.WatchOpts, sink chan<- *FlashtestationRegistryTEEServiceRegistered) (event.Subscription, error) {

	logs, sub, err := _FlashtestationRegistry.contract.WatchLogs(opts, "TEEServiceRegistered")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FlashtestationRegistryTEEServiceRegistered)
				if err := _FlashtestationRegistry.contract.UnpackLog(event, "TEEServiceRegistered", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseTEEServiceRegistered is a log parse operation binding the contract event 0x206fdb1a74851a8542447b8b6704db24a36b906a7297cc23c2b984dc357b9978.
//
// Solidity: event TEEServiceRegistered(address teeAddress, bytes rawQuote, bool alreadyExists)
func (_FlashtestationRegistry *FlashtestationRegistryFilterer) ParseTEEServiceRegistered(log types.Log) (*FlashtestationRegistryTEEServiceRegistered, error) {
	event := new(FlashtestationRegistryTEEServiceRegistered)
	if err := _FlashtestationRegistry.contract.UnpackLog(event, "TEEServiceRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// FlashtestationRegistryUpgradedIterator is returned from FilterUpgraded and is used to iterate over the raw logs and unpacked data for Upgraded events raised by the FlashtestationRegistry contract.
type FlashtestationRegistryUpgradedIterator struct {
	Event *FlashtestationRegistryUpgraded // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *FlashtestationRegistryUpgradedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FlashtestationRegistryUpgraded)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(FlashtestationRegistryUpgraded)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *FlashtestationRegistryUpgradedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FlashtestationRegistryUpgradedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FlashtestationRegistryUpgraded represents a Upgraded event raised by the FlashtestationRegistry contract.
type FlashtestationRegistryUpgraded struct {
	Implementation common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterUpgraded is a free log retrieval operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_FlashtestationRegistry *FlashtestationRegistryFilterer) FilterUpgraded(opts *bind.FilterOpts, implementation []common.Address) (*FlashtestationRegistryUpgradedIterator, error) {

	var implementationRule []interface{}
	for _, implementationItem := range implementation {
		implementationRule = append(implementationRule, implementationItem)
	}

	logs, sub, err := _FlashtestationRegistry.contract.FilterLogs(opts, "Upgraded", implementationRule)
	if err != nil {
		return nil, err
	}
	return &FlashtestationRegistryUpgradedIterator{contract: _FlashtestationRegistry.contract, event: "Upgraded", logs: logs, sub: sub}, nil
}

// WatchUpgraded is a free log subscription operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_FlashtestationRegistry *FlashtestationRegistryFilterer) WatchUpgraded(opts *bind.WatchOpts, sink chan<- *FlashtestationRegistryUpgraded, implementation []common.Address) (event.Subscription, error) {

	var implementationRule []interface{}
	for _, implementationItem := range implementation {
		implementationRule = append(implementationRule, implementationItem)
	}

	logs, sub, err := _FlashtestationRegistry.contract.WatchLogs(opts, "Upgraded", implementationRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FlashtestationRegistryUpgraded)
				if err := _FlashtestationRegistry.contract.UnpackLog(event, "Upgraded", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseUpgraded is a log parse operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_FlashtestationRegistry *FlashtestationRegistryFilterer) ParseUpgraded(log types.Log) (*FlashtestationRegistryUpgraded, error) {
	event := new(FlashtestationRegistryUpgraded)
	if err := _FlashtestationRegistry.contract.UnpackLog(event, "Upgraded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
