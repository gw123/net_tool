package types

type TypeClientId uint64

const ConnectTypeAsServer = 1
const ConnectTypeAsClient = 2

type LoginServer struct {
	ClientId    TypeClientId
	Timestamp   uint32
	ConnectType uint32
}


type LoginClientServer struct {
	AuthCode       uint64
	LoginServer
	RemoteClientId TypeClientId
}

type ServerClient struct {
	ServerClientId    TypeClientId
	CommanderClientId TypeClientId
	Ip                uint32
	Port              uint16
	IsBusy            uint8
}

type TypeClientMap map[TypeClientId]ServerClient
