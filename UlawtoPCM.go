package main

import (
  "encoding/binary"
  "fmt"
  "io"
  "os"
)

// µ-law decoding table
var muLawTable [256]int16

// Initialize the µ-law to PCM conversion table
func createMuLawTable() {
  for i := 0; i < 256; i++ {
    muLaw := ^uint8(i)     // Invert the bits
    sign := muLaw & 0x80   // Extracting Sign bit
    exponent := (muLaw >> 4) & 0x07 // Extracting Exponent by shifting right 4 bits
    mantissa := muLaw & 0x0F //  Extracting Mantissa by masking with 0x0F (15 in decimal)
    magnitude := ((int(mantissa) << 3) + 0x84) << exponent // Calculating magnitude
    if sign != 0 {
      magnitude = -magnitude // Applying sign bit
    }
    muLawTable[i] = int16(magnitude) // Storing the value in the table
  }
}

// Convert a µ-law byte to 16-bit PCM sample
func muLawToPCM(b byte) int16 {
  return muLawTable[b]
}

// Write WAV header to the output file
func createWavHeader(w io.Writer, dataSize int) error {
  var header [44]byte

  // "RIFF" (Resource Interchange File Format)
  copy(header[0:], []byte("RIFF"))
  binary.LittleEndian.PutUint32(header[4:], uint32(36+dataSize))

  // "WAVE"
  copy(header[8:], []byte("WAVE"))

  // "fmt "
  copy(header[12:], []byte("fmt "))
  binary.LittleEndian.PutUint32(header[16:], 16)        // Subchunk1Size
  binary.LittleEndian.PutUint16(header[20:], 1)         // PCM format
  binary.LittleEndian.PutUint16(header[22:], 1)         // Mono
  binary.LittleEndian.PutUint32(header[24:], 16000)     // Sample rate
  binary.LittleEndian.PutUint32(header[28:], 16000*2)   // Byte rate
  binary.LittleEndian.PutUint16(header[32:], 2)         // Block align
  binary.LittleEndian.PutUint16(header[34:], 16)        // Bits per sample

  // "data"
  copy(header[36:], []byte("data"))
  binary.LittleEndian.PutUint32(header[40:], uint32(dataSize))

  _, err := w.Write(header[:])
  return err
}

func main() {
  // Creating the µ-law decoding table
  createMuLawTable()

  // Step 1: Open and read input.ulaw
  input, err := os.Open("input.ulaw")
  if err != nil {
    fmt.Println("Failed to open input.ulaw:", err)
    return
  }
  defer input.Close()

  rawData, err := io.ReadAll(input)
  if err != nil {
    fmt.Println("Failed to read input.ulaw:", err)
    return
  }

  // Step 2: Convert µ-law to 16-bit PCM with upsampling
  var pcmData []int16
  // Processing 160-byte chunks
 chunkSize := 160
 for i := 0; i < len(rawData); i += chunkSize {
  end := i + chunkSize
  if end > len(rawData) {
   end = len(rawData)
  }
  chunk := rawData[i:end]

  for _, b := range chunk {
   pcm := muLawToPCM(b)

   // Upsample 8kHz ➜ 16kHz 
   pcmData = append(pcmData, pcm, pcm)
  }
}
  /*for _, b := range rawData {
    sample := muLawToPCM(b)

    // Upsample from 8kHz ➜ 16kHz 
    pcmData = append(pcmData, sample, sample)*/
  

  // Step 3: Create output WAV file
  output, err := os.Create("output.wav")
  if err != nil {
    fmt.Println("Failed to create output.wav:", err)
    return
  }
  defer output.Close()

  dataSize := len(pcmData) * 2 // 2 bytes per sample
  err = createWavHeader(output, dataSize)
  if err != nil {
    fmt.Println("Failed to write WAV header:", err)
    return
  }

  // Step 4: Write PCM samples
  for _, sample := range pcmData {
    err := binary.Write(output, binary.LittleEndian, sample)
    if err != nil {
      fmt.Println("Failed to write PCM data:", err)
      return
    }
  }

  fmt.Println("output.wav created successfully ")
}
