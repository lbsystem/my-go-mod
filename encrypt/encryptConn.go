package encrypt

import (
	
	"net"
)

type EncryptListener struct{
	net.Listener
	Ecdhe *ECDHEManager
}

func (l *EncryptListener)Accept() (net.Conn, error){
	buf:=make([]byte,1024)
	c, err := l.Listener.Accept()
	if err!=nil{
		return nil,err
	}
	n, err := c.Read(buf)
	if err!=nil{
		return nil,err
	}
	c.Write(l.Ecdhe.PublicKeyBytes())
	key, err := l.Ecdhe.ComputeSharedSecret(buf[:n])
	if err!=nil{
		return nil,err
	}
	crtConn, err := NewCTRConn(c,key, key[16:],key[16:])
	if err!=nil{
		return nil,err
	}
	return crtConn,nil
}


func NewEncrypListener(ip,ECDHE_MODE string)(*EncryptListener,error){
	
	ecdhe, err := NewECDHEManager(ECDHE_MODE)
	if err!=nil{
		return nil,err
	}

	l, err := net.Listen("tcp", ip)
	if err!=nil{
		return nil,err
	}
	
	e:=new(EncryptListener)
	e.Ecdhe=ecdhe
	if err!=nil{
		return nil,err
	}
	e.Listener=l
	return e,nil
	
}



type EncryptTcpConn struct{
	net.Conn
	Ecdhe *ECDHEManager
}


func NewEncryptTcpConn(host,ECDHE_MODE string)(*EncryptTcpConn,error){
	buf:=make([]byte,1024)
	ecdhe, err := NewECDHEManager(ECDHE_MODE)
	if err!=nil{
		return nil,err
	}
	e:=new(EncryptTcpConn)
	e.Ecdhe=ecdhe
	c, err := net.Dial("tcp", host)
	if err!=nil{
		return nil,err
	}
	c.Write(ecdhe.PublicKeyBytes())
	n, err := c.Read(buf)
	if err!=nil{
		return nil,err
	}
	key, err := ecdhe.ComputeSharedSecret(buf[:n])
	if err!=nil{
		return nil,err
	}
	
	crtConn, err := NewCTRConn(c, key,key[16:], key[16:])
	if err!=nil{
		return nil,err
	}
	e.Conn=crtConn
	return e,nil
}
