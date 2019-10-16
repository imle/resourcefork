package resourcefork

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

//noinspection GoUnusedConst,GoNameStartsWithPackageName
const ResourceForkIDOffset = 128

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

	rdOffset := offsetData + lengthData
	resourcesData := fileBytes[offsetData:rdOffset:rdOffset]
	rmOffset := offsetMap + lengthMap
	resourcesMap := fileBytes[offsetMap:rmOffset:rmOffset]

	offsetTypeList := binary.BigEndian.Uint16(resourcesMap[24:])
	offsetNameList := binary.BigEndian.Uint16(resourcesMap[26:])

	typeList := resourcesMap[offsetTypeList:offsetNameList:offsetNameList]
	nameList := resourcesMap[offsetNameList:] // Goes to end of buffer

	numberOfTypes := (binary.BigEndian.Uint16(typeList[0:]) + 1) & 0xFFFF

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
			if offsetResourceName == 0xFFFF {
				resourceName = ""
			} else {
				lengthResourceName := nameList[offsetResourceName]

				resourceName = decodeMacRoman(nameList[offsetResourceName+1 : offsetResourceName+1+uint16(lengthResourceName)])
			}

			resourceDataOffsetMSB := uint32(typeList[offset+12*j+5])
			resourceDataOffsetLSB := uint32(binary.BigEndian.Uint16(typeList[offset+12*j+6:]))

			resourceDataOffset := resourceDataOffsetMSB<<16 + resourceDataOffsetLSB
			resourceDataLength := binary.BigEndian.Uint32(resourcesData[resourceDataOffset:])

			rdOffset := resourceDataOffset + 4 + resourceDataLength
			resourceData := resourcesData[resourceDataOffset+4 : rdOffset : rdOffset]

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

func getNovaDataFilesFromPath(path string) (filePaths []string, err error) {
	path, err = filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if fi.IsDir() {
		paths, err := ioutil.ReadDir(path)
		if err != nil {
			return nil, err
		}

		for _, v := range paths {
			if !v.IsDir() && filepath.Ext(v.Name()) != ".ndat" {
				continue
			}

			fromPath, err := getNovaDataFilesFromPath(filepath.Join(path, v.Name()))
			if err != nil {
				return nil, err
			}

			filePaths = append(filePaths, fromPath...)
		}

		return filePaths, nil

	} else if filepath.Ext(fi.Name()) == ".ndat" {
		return append(filePaths, path), nil
	}

	return nil, nil
}

func ReadResourceForkFromPath(paths ...string) (*ResourceFork, error) {
	var filePaths []string
	for _, v := range paths {
		fromPath, err := getNovaDataFilesFromPath(v)
		if err != nil {
			return nil, err
		}

		filePaths = append(filePaths, fromPath...)
	}

	resFork := &ResourceFork{Resources: map[string]map[uint16]Resource{}}
	for _, v := range filePaths {
		dat, err := ioutil.ReadFile(v)
		if err != nil {
			return nil, err
		}

		rf, err := ReadResourceForkFromBytes(dat)
		if err != nil {
			return nil, err
		}

		for name, resMap := range rf.Resources {
			for idx, res := range resMap {
				if _, ok := resFork.Resources[name]; !ok {
					resFork.Resources[name] = map[uint16]Resource{}
				}

				resFork.Resources[name][idx] = res
			}
		}
	}

	delete(resFork.Resources, "csüm") // Don't care about checksum
	delete(resFork.Resources, "dsïg") // Don't care about digital signature

	return resFork, nil
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
