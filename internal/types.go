package internal

import (
	"crypto/sha256"
	"fmt"
)

//Block defines blockchain block structure
type Message struct{
	Issuer int 
	Round int 
	Time int64 
	Payload []byte
}

func (m Message) Hash() []byte{
	hash := sha256.New()
	//writing paylaod first seems more memory efficient because 
	//no need to create a long string
	_, err := hash.Write(m.Payload)
	if err != nil{
		panic(err)
	}

	str := fmt.Sprintf("%d, %d,", m.Issuer, m.Round)
	_, err = hash.Write([]byte(str))
	if err != nil{
		panic(err)
	}
	return hash.Sum(nil)
}

//BlockChunk defies a chunk of a block.
//Blockchunks dissemiate fater in the gossip network because they are very small compared to a Block
type Chunk struct{
	//NodeID of the issuer 
	Issuer int 
	//Round of the Block 
	Round int
	//Time of block creation 
	Time int64
	//..The number of expected chunnks
	ChunkCount int
	//Chunk index 
	ChunkIndex int
	//Contains merkel path to authentication teh chunk 
	MerklePath []byte 
	//Chunk Payload 
	Payload []byte
}

//Hash produces the digest of a Blocckhain it consider all fields of a Blockchain 
func (c Chunk) Hash() []byte{
	hash := sha256.New()
	//
	_, err := hash.Write(c.Payload)
	if err != nil{
		panic(err)
	}
	str := fmt.Sprintf("%d,%d,%d,%d", c.Round, c.ChunkCount, c.ChunkIndex, c.Issuer)
	_, err = hash.Write([]byte(str))
	if err != nil {
		panic(err)
	}

	return hash.Sum(nil)

}
