# fabric-relayer

## replayer目录结构
`
drwxr-xr-x 8 root root     4096 11月  2 10:51 .

drwxr--r-- 2 root root     4096 11月  2 10:40 Log

-rwxr-xr-x 1 root root 39631216 11月  2 10:40 main

drwxr-xr-x 2 root root     4096 11月  2 10:40 wallet

drwxr-xr-x 4 root root     4096 11月  2 10:40 peerOrganizations

drwxr-xr-x 3 root root     4096 11月  2 10:40 ordererOrganizations

drwxr-xr-x 2 root root     4096 11月  2 10:40 boltdb

drwxr-xr-x 2 root root     4096 11月  2 10:40 config

-rw-r--r-- 1 root root       16 11月  2 10:40 README.md

drwxr-xr-x 4 root root     4096 11月  2 10:40 ..
`
+ Log: relayer日志输出目录
+ main： realyer应用程序
+ wallet: poly的钱包文件
+ peerOrganizations&ordererOrganizations: peer节点和order节点证书
+ boltdb: 本地数据库，不需要额外安装数据库服务
+ config: relayer的全局配置
+ README.md: relayer的介绍和使用说明

## 配置

配置包括三部分，relayer全局配置、fabric环境配置、fabric账户配置

### relayer全局配置

配置在文件config/config.json

`
{
  "PolyConfig": {
    "RestURL": "http://106.75.226.11:40336",
    "EntranceContractAddress": "0300000000000000000000000000000000000000",
    "WalletFile": "./wallet/wallet.dat",
    "WalletPwd": "4cUYqGj2yib718E7ZmGQc"
  },
  "FabricConfig": {
    "SideChainId": 7,
    "BlockConfig": 1,
    "Channel":"mychannel",
    "Chaincode": "ccm1"
  },
  "BoltDbPath": "./boltdb",
}
`
relayer全局配置包括三部分：

1、配置poly链的信息

+ RestURL: poly节点url
+ WalletFile&WalletPwd: poly的钱包文件

2、fabric链信息

+ Channel: 跨链管理合约部署的channel
+ Chaincode: 跨链管理合约的链码ID

3、bolt数据库
+ BoltDbPath: bolt数据库文件存放路径

### fabric环境配置

配置在文件config/config_e2e.yaml

relayer启动后会将环境变量FABRIC_RELAYER_PATH设置为当前目录，需要用户根据自己的fabric环境进行以下配置：

1、relayer连接fabric的TLS连接配置
`
    client:
      key:
        path: ${FABRIC_RELAYER_PATH}/peerOrganizations/org1.example.com/users/User1@org1.example.com/tls/client.key
      cert:
        path: ${FABRIC_RELAYER_PATH}/peerOrganizations/org1.example.com/users/User1@org1.example.com/tls/client.crt
`

2、fabric的peer信息

`
peers:
  peer0.org1.example.com:
    # this URL is used to send endorsement and query requests
    # [Optional] Default: Infer from hostname
    url: peer0.org1.example.com:7051

    grpcOptions:
      ssl-target-name-override: peer0.org1.example.com
      # These parameters should be set in coordination with the keepalive policy on the server,
      # as incompatible settings can result in closing of connection.
      # When duration of the 'keep-alive-time' is set to 0 or less the keep alive client parameters are disabled
      keep-alive-time: 0s
      keep-alive-timeout: 20s
      keep-alive-permit: false
      fail-fast: false
      # allow-insecure will be taken into consideration if address has no protocol defined, if true then grpc or else grpcs
      allow-insecure: false

    tlsCACerts:
      # Certificate location absolute path
      path: ${FABRIC_RELAYER_PATH}/peerOrganizations/org1.example.com/tlsca/tlsca.org1.example.com-cert.pem

  peer0.org2.example.com:
    # this URL is used to send endorsement and query requests
    # [Optional] Default: Infer from hostname
    url: peer0.org2.example.com:9051

    grpcOptions:
      ssl-target-name-override: peer0.org2.example.com
      # These parameters should be set in coordination with the keepalive policy on the server,
      # as incompatible settings can result in closing of connection.
      # When duration of the 'keep-alive-time' is set to 0 or less the keep alive client parameters are disabled
      keep-alive-time: 0s
      keep-alive-timeout: 20s
      keep-alive-permit: false
      fail-fast: false
      # allow-insecure will be taken into consideration if address has no protocol defined, if true then grpc or else grpcs
      allow-insecure: false

    tlsCACerts:
      # Certificate location absolute path
      path: ${FABRIC_RELAYER_PATH}/peerOrganizations/org2.example.com/tlsca/tlsca.org2.example.com-cert.pem
`

3、fabric的组织信息

`
organizations:
  Org1:
    mspid: Org1MSP
    peers:
      - peer0.org1.example.com

    cryptoPath:  peerOrganizations/org1.example.com/users/{username}@org1.example.com/msp

`

4、fabric的order节点信息
`
orderers:
  orderer.example.com:
    # [Optional] Default: Infer from hostname
    url: orderer.example.com:7050

    # these are standard properties defined by the gRPC library
    # they will be passed in as-is to gRPC client constructor
    grpcOptions:
      ssl-target-name-override: orderer.example.com
      # These parameters should be set in coordination with the keepalive policy on the server,
      # as incompatible settings can result in closing of connection.
      # When duration of the 'keep-alive-time' is set to 0 or less the keep alive client parameters are disabled
      keep-alive-time: 0s
      keep-alive-timeout: 20s
      keep-alive-permit: false
      fail-fast: false
      # allow-insecure will be taken into consideration if address has no protocol defined, if true then grpc or else grpcs
      allow-insecure: false

    tlsCACerts:
      # Certificate location absolute path
      path: ${FABRIC_RELAYER_PATH}/ordererOrganizations/example.com/tlsca/tlsca.example.com-cert.pem
`


## 启动relayer

`
nohup ./main &
`