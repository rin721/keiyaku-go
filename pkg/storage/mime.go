package storage

import (
	"github.com/gabriel-vasile/mimetype"
)

// DetectMIME detects the content type of a file.
func (s *Storage) DetectMIME(name string) (MIMEInfo, error) {
	cleaned, err := cleanStoragePath(name)
	if err != nil {
		return MIMEInfo{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	file, err := s.fs.Open(cleaned)
	if err != nil {
		return MIMEInfo{}, err
	}
	defer file.Close()

	detected, err := mimetype.DetectReader(file)
	if err != nil {
		return MIMEInfo{}, err
	}
	return mimeInfo(detected), nil
}

// DetectBytes detects the content type of data.
func DetectBytes(data []byte) MIMEInfo {
	return mimeInfo(mimetype.Detect(data))
}

func mimeInfo(detected *mimetype.MIME) MIMEInfo {
	if detected == nil {
		return MIMEInfo{}
	}
	return MIMEInfo{
		MIME:      detected.String(),
		Extension: detected.Extension(),
	}
}
