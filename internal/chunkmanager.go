package internal

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"math"
	"time"

	"github.com/klauspost/reedsolomon"
)

func ChunkMessage(message Message, numberOfChunks int, dataChunkCount int) []Chunk {

	blockBytes := encodeToBytes(message)
	chunks := constructChunks(message, blockBytes, numberOfChunks, dataChunkCount)
	return chunks
}

// mergeChunks assumes that sanity checks are done before calling this function
func MergeChunks(chunks []Chunk, numberOfChunks int, dataChunkCount int) Message {

	data := make([][]byte, numberOfChunks)
	for i := 0; i < len(chunks); i++ {
		c := chunks[i]
		data[c.ChunkIndex] = c.Payload
	}

	////////////////// Reed-Solomon Erasure Coding  //////////////////
	start := time.Now()
	parityChunkCount := numberOfChunks - dataChunkCount

	enc, err := reedsolomon.New(dataChunkCount, parityChunkCount)
	if err != nil {
		panic(err)
	}

	err = enc.ReconstructData(data)
	if err != nil {
		panic(err)
	}
	log.Printf("[Decode] Erasure coding took %s", time.Since(start))
	///////////////////////////////////////////////////////////////////

	// combines data
	// TODO: is it a good way to implement this
	var blockData []byte
	for i := 0; i < dataChunkCount; i++ {
		blockData = append(blockData, data[i]...)
	}
	//

	return decodeToBlock(blockData)
}

func constructChunks(message Message, blockBytes []byte, numberOfChunks int, dataChunkCount int) []Chunk {

	var chunks []Chunk

	////////////////// Reed-Solomon Erasure Coding  //////////////////
	start := time.Now()
	parityChunkCount := numberOfChunks - dataChunkCount

	enc, err := reedsolomon.New(dataChunkCount, parityChunkCount)
	if err != nil {
		panic(err)
	}

	data, err := enc.Split(blockBytes)
	if err != nil {
		panic(err)
	}

	err = enc.Encode(data)
	if err != nil {
		panic(err)
	}
	log.Printf("[Encode] Erasure coding took %s", time.Since(start))
	///////////////////////////////////////////////////////////////////

	// Size of a Merkle path in Bytes to authenticate the chunk
	merklePathSize := int(math.Log2(float64(numberOfChunks))+1) * sha256.Size
	log.Printf("Merkle path size is %d bytes\n", merklePathSize)

	for i := 0; i < numberOfChunks; i++ {

		chunk := Chunk{
			Issuer:     message.Issuer,
			Round:      message.Round,
			Time:       message.Time,
			ChunkCount: numberOfChunks,
			ChunkIndex: i,
			MerklePath: GetRandomBySlices(merklePathSize),
			Payload:    data[i],
		}

		chunks = append(chunks, chunk)
	}

	return chunks
}

// https://gist.github.com/SteveBate/042960baa7a4795c3565
func encodeToBytes(p interface{}) []byte {

	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(p)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func decodeToBlock(data []byte) Message {

	message := Message{}
	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&message)
	if err != nil {
		panic(err)
	}
	return message
}
