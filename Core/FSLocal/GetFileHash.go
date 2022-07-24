package FSLocal

import "Archcopy/HashFile"

func (o *FSAccessLocalStruct) GetFileHash(Source string) (string, error) {
	return HashFile.HashFile(Source, nil)
}
