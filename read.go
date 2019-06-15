package resourcefork

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io/ioutil"
)

const highCharsUnicode = "ÄÅÇÉÑÖÜáàâäãåçéèêëíìîïñóòôöõúùûü†°¢£§•¶ß®©™´¨≠ÆØ∞±≤≥¥µ∂∑∏π∫ªºΩæø" +
	"¿¡¬√ƒ≈∆«»… ÀÃÕŒœ–—“”‘’÷◊ÿŸ⁄€‹›ﬁﬂ‡·‚„‰ÂÊÁËÈÍÎÏÌÓÔÒÚÛÙıˆ˜¯˘˙˚¸˝˛ˇ"

var highCharsUnicodeRunes = bytes.Runes([]byte(highCharsUnicode))

type Resource struct {
	Type string
	ID   uint16
	Name string
	Data []byte
}

type ResourceFork struct {
	Resources map[string]map[uint16]Resource
}

func ReadResourceForkFromBytes(fileBytes []byte) (r *ResourceFork, err error) {
	defer func() {
		if r := recover(); r != nil {
			//fmt.Println("Recovered in f", r)
			// find out exactly what the error was and set err
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("unknown panic")
			}

			r = nil
		}
	}()

	resources := ResourceFork{
		Resources: map[string]map[uint16]Resource{},
	}

	offsetData := binary.BigEndian.Uint32(fileBytes[0*4:])
	offsetMap := binary.BigEndian.Uint32(fileBytes[1*4:])
	lengthData := binary.BigEndian.Uint32(fileBytes[2*4:])
	lengthMap := binary.BigEndian.Uint32(fileBytes[3*4:])

	// Will panic if invalid
	_ = offsetData != binary.BigEndian.Uint32(fileBytes[offsetData:])
	_ = offsetMap != binary.BigEndian.Uint32(fileBytes[offsetMap:])
	_ = lengthData != binary.BigEndian.Uint32(fileBytes[lengthData:])
	_ = lengthMap != binary.BigEndian.Uint32(fileBytes[lengthMap:])

	resourcesData := fileBytes[offsetData : offsetData+lengthData]
	resourcesMap := fileBytes[offsetMap : offsetMap+lengthMap]

	offsetTypeList := binary.BigEndian.Uint16(resourcesMap[24:])
	offsetNameList := binary.BigEndian.Uint16(resourcesMap[26:])

	typeList := resourcesMap[offsetTypeList:offsetNameList]
	nameList := resourcesMap[offsetNameList:] // Goes to end of buffer

	numberOfTypes := (binary.BigEndian.Uint16(typeList[0:]) + 1) & 0xffff

	var i uint16
	for i = 0; i < numberOfTypes; i++ {
		resourceType := decodeMacRoman(typeList[2+8*i : 2+8*i+4])

		if _, ok := resources.Resources[resourceType]; ok {
			return nil, errors.New("duplicate resource type")
		}

		quantity := binary.BigEndian.Uint16(typeList[6+8*i:]) + 1
		offset := binary.BigEndian.Uint16(typeList[8+8*i:])

		resources.Resources[resourceType] = map[uint16]Resource{}

		var j uint16
		for j = 0; j < quantity; j++ {
			resourceID := binary.BigEndian.Uint16(typeList[offset+12*j:])
			offsetResourceName := binary.BigEndian.Uint16(typeList[offset+12*j+2:])

			var resourceName string
			if offsetResourceName == 0xffff {
				resourceName = ""
			} else {
				lengthResourceName := uint8(nameList[offsetResourceName])

				resourceName = decodeMacRoman(nameList[offsetResourceName+1 : offsetResourceName+1+uint16(lengthResourceName)])
			}

			resourceDataOffsetMSB := uint32(typeList[offset+12*j+5])
			resourceDataOffsetLSB := uint32(binary.BigEndian.Uint16(typeList[offset+12*j+6:]))

			resourceDataOffset := resourceDataOffsetMSB<<16 + resourceDataOffsetLSB
			resourceDataLength := binary.BigEndian.Uint32(resourcesData[resourceDataOffset:])

			resourceData := resourcesData[resourceDataOffset+4 : resourceDataOffset+4+resourceDataLength]

			resources.Resources[resourceType][resourceID] = Resource{
				Type: resourceType,
				ID:   resourceID,
				Name: resourceName,
				Data: make([]byte, len(resourceData)),
			}

			copy(resources.Resources[resourceType][resourceID].Data, resourceData)
		}
	}

	return &resources, nil
}

func ReadResourceForkFromPath(p string) (*ResourceFork, error) {
	dat, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}

	return ReadResourceForkFromBytes(dat)
}

// https://gist.github.com/jrus/3113240
func decodeMacRoman(macRomanByteString []byte) string {
	returnString := ""
	for _, b := range macRomanByteString {
		if b < 0x80 {
			returnString += string(rune(b))
		} else {
			returnString += string(highCharsUnicodeRunes[b-0x80])
		}
	}

	return returnString
}
