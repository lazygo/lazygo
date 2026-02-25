package utils

import (
	"fmt"
	"testing"
)

func TestGenerateOrderNo(t *testing.T) {
	orderNo := GenerateSerialNumber(1000)
	fmt.Println(orderNo)
	orderNo = GenerateSerialNumber(1000)
	fmt.Println(orderNo)
	orderNo = GenerateSerialNumber(1000)
	fmt.Println(orderNo)
	orderNo = GenerateSerialNumber(1000)
	fmt.Println(orderNo)
	orderNo = GenerateSerialNumber(1000)
	fmt.Println(orderNo)
	orderNo = GenerateSerialNumber(1000)
	fmt.Println(orderNo)
	orderNo = GenerateSerialNumber(1000)
	fmt.Println(orderNo)
	orderNo = GenerateSerialNumber(1000)
	fmt.Println(orderNo)
	orderNo = GenerateSerialNumber(1000)
	fmt.Println(orderNo)
	orderNo = GenerateSerialNumber(1000)
	fmt.Println(orderNo)
	orderNo = GenerateSerialNumber(1000)
	fmt.Println(orderNo)
	orderNo = GenerateSerialNumber(1000)
	fmt.Println(orderNo)
	orderNo = GenerateSerialNumber(1000)
	fmt.Println(orderNo)
	orderNo = GenerateSerialNumber(1000)
	fmt.Println(orderNo)
}

func TestEncodeUint64(t *testing.T) {
	encoded := EncodeUint64(10000)
	if encoded != "3yR" {
		t.Error("enode fail", encoded)
	}
	decoded := DecodeUint64(encoded)
	if decoded != 10000 {
		t.Error("decode fail", decoded)
	}
}

func TestMask(t *testing.T) {
	str := Mask("123", 1, 1)
	if str != "1*3" {
		t.Error("mask fail", str)
	}
	str = Mask("123", 1, 3)
	if str != "123" {
		t.Error("mask fail", str)
	}
	str = Mask("123", 1, 4)
	if str != "123" {
		t.Error("mask fail", str)
	}
	str = Mask("123", 1, 0)
	if str != "1**" {
		t.Error("mask fail", str)
	}
	str = Mask("123", 1, 0)
	if str != "1**" {
		t.Error("mask fail", str)
	}
	str = Mask("18500230535", 2, 3)
	if str != "18******535" {
		t.Error("mask fail", str)
	}
	str = Mask("lzp9421@qq.com", 2, 3)
	if str != "lz******com" {
		t.Error("mask fail", str)
	}
	str = Mask("", 2, 3)
	if str != "" {
		t.Error("mask fail", str)
	}
}
