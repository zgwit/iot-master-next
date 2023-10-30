package plugin

import (
	"encoding/binary"
	"errors"
	"math"
)

const (
	MODBUS_FUNC_COIL_RW     = "coil_rw"
	MODBUS_FUNC_COIL_R      = "coil_r"
	MODBUS_FUNC_REGISTER_RW = "register_rw"
	MODBUS_FUNC_REGISTER_R  = "register_r"
)

var MODBUS_FUNC = []string{MODBUS_FUNC_COIL_RW, MODBUS_FUNC_COIL_R, MODBUS_FUNC_REGISTER_RW, MODBUS_FUNC_REGISTER_R}

const (
	MODBUS_DT_BOOL   = "bool"
	MODBUS_DT_UINT16 = "uint16"
	MODBUS_DT_INT16  = "int16"
	MODBUS_DT_UINT32 = "uint32"
	MODBUS_DT_INT32  = "int32"
	MODBUS_DT_FLOAT  = "float"
)

const (
	MODBUS_ED_BIG    = "big"
	MODBUS_ED_LITTLE = "little"
)

type ModbusPoint struct {
	Function string `form:"function" bson:"function" json:"function"`
	Address  uint16 `form:"address" bson:"address" json:"address"`
	DataType string `form:"data_type" bson:"data_type" json:"data_type"`

	value *float64
}

func (modbus_point ModbusPoint) Value() *float64 {
	return modbus_point.value
}

func calculate_crc(data []byte) (crc uint16) {

	crc = 0xFFFF

	for i := 0; i < len(data); i++ {
		crc ^= uint16(data[i])
		for j := 0; j < 8; j++ {
			if (crc & 0x0001) != 0 {
				crc >>= 1
				crc ^= 0xA001
			} else {
				crc >>= 1
			}
		}
	}

	return
}

func ModbusRtuReadByte(slave_id uint8, function string, address, quantity uint16) (result []byte) {

	result = []byte{byte(slave_id)}

	switch function {

	case MODBUS_FUNC_COIL_RW:
		result = append(result, 0x01)

	case MODBUS_FUNC_COIL_R:
		result = append(result, 0x02)

	case MODBUS_FUNC_REGISTER_RW:
		result = append(result, 0x03)

	case MODBUS_FUNC_REGISTER_R:
		result = append(result, 0x04)
	}

	result = append(result, byte(address/0x100))
	result = append(result, byte(address%0x100))

	result = append(result, byte(quantity/0x100))
	result = append(result, byte(quantity%0x100))

	crc := calculate_crc(result)

	result = append(result, byte(crc%0x100))
	result = append(result, byte(crc/0x100))

	return
}

func check_crc(data []byte, crc []byte) (result bool) {

	data_crc := calculate_crc(data)

	return byte(data_crc/0x100) == crc[1] && byte(data_crc%0x100) == crc[0]
}

func reverseBytes(input []byte) []byte {

	reversed := make([]byte, len(input))

	for i := 0; i < len(input); i++ {
		reversed[i] = input[len(input)-1-i]
	}

	return reversed
}

func bool_to_float64(value bool) float64 {

	if value {
		return 1
	} else {
		return 0
	}
}

func float32_to_byte(float float32) []byte {
	bits := math.Float32bits(float)
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, bits)

	return bytes
}

func byte_to_float32(bytes []byte) float32 {
	bits := binary.LittleEndian.Uint32(bytes)

	return math.Float32frombits(bits)
}

func ModbusRtuGetData(result []byte, slave_id uint8, address uint16, endian string, points map[string]*ModbusPoint) (err error) {

	if len(result) < 6 || int(result[2]) != len(result)-5 || result[2]%2 != 0 {
		err = errors.New("wrong format")
		return
	}

	_slave_id, function, data_byte, crc := result[0], MODBUS_FUNC[result[1]], result[3:3+int(result[2])], result[len(result)-2:]

	if !check_crc(result[:len(result)-2], crc) {
		err = errors.New("wrong crc")
		return
	}

	if _slave_id != slave_id {
		err = errors.New("wrong slave_id")
		return
	}

	for _, point := range points {
		if point.Function == function {
			point.value = nil
		}
	}

	switch function {

	case MODBUS_FUNC_COIL_RW, MODBUS_FUNC_COIL_R:

		for _, item := range data_byte {

			for wei := 7; wei >= 0; wei-- {

				for _, point := range points {

					value := bool_to_float64(uint16(item&(0x01<<wei)) != 0)

					if point.Function == function || point.Address == address {
						point.value = &value
					}
				}

				address++
			}
		}

	case MODBUS_FUNC_REGISTER_RW, MODBUS_FUNC_REGISTER_R:

		for wei := 0; wei < len(data_byte); wei += 2 {

			for _, point := range points {

				if point.Function == function || point.Address == address {

					_data_byte := []byte{}

					switch point.DataType {

					case MODBUS_DT_BOOL, MODBUS_DT_UINT16, MODBUS_DT_INT16:

						if wei+1 > len(data_byte)-1 {
							continue
						}

						_data_byte = append(_data_byte, data_byte[wei+0], data_byte[wei+1])

					case MODBUS_DT_UINT32, MODBUS_DT_INT32, MODBUS_DT_FLOAT:

						if wei+3 > len(data_byte)-1 {
							continue
						}

						_data_byte = append(_data_byte, data_byte[wei+0], data_byte[wei+1])
						_data_byte = append(_data_byte, data_byte[wei+2], data_byte[wei+3])

					default:
						continue
					}

					switch endian {
					case MODBUS_ED_BIG:
					case MODBUS_ED_LITTLE:
						_data_byte = reverseBytes(_data_byte)
					default:
						continue
					}

					value := float64(0)

					switch point.DataType {
					case MODBUS_DT_BOOL:
						value = bool_to_float64(_data_byte[1] != 0)
					case MODBUS_DT_UINT16:
						value = float64(uint16(_data_byte[0])*0x100 + uint16(_data_byte[1]))
					case MODBUS_DT_INT16:
						value = float64(int16(uint16(_data_byte[0])*0x100 + uint16(_data_byte[1])))
					case MODBUS_DT_UINT32:
						value = float64(uint32(_data_byte[0])*0x1000000 + uint32(_data_byte[1])*0x10000 + uint32(_data_byte[2])*0x100 + uint32(_data_byte[3]))
					case MODBUS_DT_INT32:
						value = float64(int32(uint32(_data_byte[0])*0x1000000 + uint32(_data_byte[1])*0x10000 + uint32(_data_byte[2])*0x100 + uint32(_data_byte[3])))
					case MODBUS_DT_FLOAT:
						value = float64(byte_to_float32(_data_byte))
					}

					point.value = &value
				}
			}

			address++
		}

	default:
		err = errors.New("wrong function code")
		return
	}

	return
}
