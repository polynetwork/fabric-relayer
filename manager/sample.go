package manager
//
//import (
//	"encoding/hex"
//	"fmt"
//
//	"github.com/hyperledger/fabric/core/chaincode/shim"
//	"github.com/hyperledger/fabric/protos/peer"
//)
//
//type Sample struct {
//}
//
//func (s *Sample) Init(stub shim.ChaincodeStubInterface) peer.Response {
//	return shim.Success(nil)
//}
//
//func (s *Sample) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
//	function, _ := stub.GetFunctionAndParameters()
//	args := stub.GetArgs()
//
//	switch function {
//	case "test":
//		return s.test(stub, args)
//	case "call":
//		return s.call(stub, args)
//	}
//
//	return shim.Error("error")
//}
//
//func (s *Sample) test(stub shim.ChaincodeStubInterface, args [][]byte) peer.Response {
//	raw, err := stub.GetCreator()
//	if err != nil {
//		return shim.Error(err.Error())
//	}
//	fmt.Println("res is" + hex.EncodeToString(raw))
//	return shim.Success(raw)
//}
//
//func (s *Sample) call(stub shim.ChaincodeStubInterface, args [][]byte) peer.Response {
//	raw, err := stub.GetCreator()
//	if err != nil {
//		return shim.Error(err.Error())
//	}
//	fmt.Println("now is" + hex.EncodeToString(raw))
//	return stub.InvokeChaincode(string(args[0]), [][]byte{[]byte("test")}, stub.GetChannelID())
//}
//
//func main() {
//	err := shim.Start(new(Sample))
//	if err != nil {
//		fmt.Println(err.Error())
//	}
//}
