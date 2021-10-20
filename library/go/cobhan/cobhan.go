package cobhan

import (
	"C"
	"encoding/json"
	"os"
	"reflect"
	"unsafe"
)
import "math"

type Buffer *C.char

const ERR_NONE = 0

//One of the provided input pointers is NULL / nil / 0
const ERR_NULL_PTR = -1

//One of the provided input buffer lengths is too large
const ERR_INPUT_BUFFER_TOO_LARGE = -2

//One of the provided output buffers was too small to receive the output
const ERR_OUTPUT_BUFFER_TOO_SMALL = -3

//Failed to copy the output into the output buffer (copy length != expected length)
const ERR_COPY_FAILED = -4

//Failed to decode a JSON input buffer
const ERR_JSON_INPUT_DECODE_FAILED = -5

//Failed to encode to JSON output buffer
const ERR_JSON_OUTPUT_ENCODE_FAILED = -6

// Reusable functions to facilitate FFI

const BUFFER_HEADER_SIZE = (64 / 8) // 64 bit buffer header provides 8 byte alignment for data pointers

var DefaultInputMaximum = math.MaxInt32

func SetDefaultInputBufferMaximum(max int) {
	DefaultInputMaximum = max
}

func bufferPtrToLength(bufferPtr unsafe.Pointer) C.int {
	return C.int(*(*int32)(bufferPtr))
}

func bufferPtrToDataPtr(bufferPtr unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(uintptr(bufferPtr) + BUFFER_HEADER_SIZE)
}

func bufferPtrToString(bufferPtr unsafe.Pointer, length C.int) string {
	dataPtr := bufferPtrToDataPtr(bufferPtr)
	return C.GoStringN((*C.char)(dataPtr), length)
}

func bufferPtrToBytes(bufferPtr unsafe.Pointer, length C.int) []byte {
	dataPtr := bufferPtrToDataPtr(bufferPtr)
	return C.GoBytes(dataPtr, length)
}

func updateBufferPtrLength(bufferPtr unsafe.Pointer, length int) {
	*(*int32)(bufferPtr) = int32(length)
}

func inputTempToBytes(ptr unsafe.Pointer, length C.int) ([]byte, int32) {
	length = 0 - length

	if DefaultInputMaximum < int(length) {
		return nil, ERR_INPUT_BUFFER_TOO_LARGE
	}

	fileName := bufferPtrToString(ptr, length)
	fileData, err := os.ReadFile(fileName)
	if err != nil {
		return nil, ERR_NULL_PTR //TODO: Temp file read error
	}
	return fileData, ERR_NONE
}

func InputBufferToBytes(srcPtr Buffer) ([]byte, int32) {
	ptr := unsafe.Pointer(srcPtr)
	length := bufferPtrToLength(ptr)

	if DefaultInputMaximum < int(length) {
		return nil, ERR_INPUT_BUFFER_TOO_LARGE
	}

	if length >= 0 {
		return bufferPtrToBytes(ptr, length), ERR_NONE
	} else {
		return inputTempToBytes(ptr, length)
	}
}

func InputBufferToString(srcPtr Buffer) (string, int32) {
	ptr := unsafe.Pointer(srcPtr)
	length := bufferPtrToLength(ptr)

	if DefaultInputMaximum < int(length) {
		return "", ERR_INPUT_BUFFER_TOO_LARGE
	}

	if length >= 0 {
		return bufferPtrToString(ptr, length), ERR_NONE
	} else {
		bytes, result := inputTempToBytes(ptr, length)
		if result < 0 {
			return "", result
		}
		return string(bytes), ERR_NONE
	}
}

func InputBufferToJson(srcPtr Buffer) (map[string]interface{}, int32) {
	bytes, result := InputBufferToBytes(srcPtr)
	if result < 0 {
		return nil, result
	}

	var loadedJson interface{}
	err := json.Unmarshal(bytes, &loadedJson)
	if err != nil {
		return nil, ERR_JSON_INPUT_DECODE_FAILED
	}
	return loadedJson.(map[string]interface{}), ERR_NONE
}

func OutputStringToBuffer(str string, dstPtr Buffer) int32 {
	return OutputBytesToBuffer([]byte(str), dstPtr)
}

func OutputJsonToBuffer(v interface{}, dstPtr Buffer) int32 {
	outputBytes, err := json.Marshal(v)
	if err != nil {
		return ERR_JSON_OUTPUT_ENCODE_FAILED
	}
	return OutputBytesToBuffer(outputBytes, dstPtr)
}

func OutputBytesToBuffer(bytes []byte, dstPtr Buffer) int32 {
	ptr := unsafe.Pointer(dstPtr)
	//Get the destination capacity from the Buffer
	dstCap := bufferPtrToLength(ptr)

	dstCapInt := int(dstCap)
	bytesLen := len(bytes)

	// Construct a byte slice out of the unsafe pointers
	var dst []byte
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&dst))
	sh.Data = (uintptr)(bufferPtrToDataPtr(ptr))
	sh.Len = dstCapInt
	sh.Cap = dstCapInt

	var result int
	if dstCapInt < bytesLen {
		// Output will not fit in supplied buffer

		// Write the data to a temp file and copy the temp file name into the buffer
		file, err := os.CreateTemp("", "cobhan-*")
		if err != nil {
			return ERR_OUTPUT_BUFFER_TOO_SMALL //TODO: Temp file write error
		}

		fileName := file.Name()
		if len(fileName) > dstCapInt {
			// Even the file path won't fit in the output buffer, we're completely out of luck now
			file.Close()
			os.Remove(fileName)
			return ERR_OUTPUT_BUFFER_TOO_SMALL
		}

		_, err = file.Write(bytes)
		if err != nil {
			file.Close()
			os.Remove(fileName)
			return ERR_OUTPUT_BUFFER_TOO_SMALL //TODO: Temp file write error
		}

		// Explicit rather than defer
		file.Close()

		fileNameBytes := ([]byte)(fileName)
		result = copy(dst, fileNameBytes)

		if result != len(fileNameBytes) {
			os.Remove(fileName)
			return ERR_COPY_FAILED
		}

		// Convert result to temp file name length
		result = 0 - result
	} else {
		result = copy(dst, bytes)

		if result != bytesLen {
			return ERR_COPY_FAILED
		}
	}

	//Update the output buffer length
	updateBufferPtrLength(ptr, result)

	return ERR_NONE
}
