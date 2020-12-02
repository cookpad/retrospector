package retrospector

import (
	"encoding/json"
	"io"

	"github.com/m-mizutani/retrospector/pkg/errors"
)

type Entity struct {
	Value
	Label      string `json:"label" dynamo:"label"`
	RecordedAt int64  `json:"recorded_at" dynamo:"recorded_at"`
}

type EntityWriter interface {
	Write(entity *Entity) (int, error)
}

type entityWriterImpl struct {
	w io.Writer
}

func (x *entityWriterImpl) Write(entity *Entity) (int, error) {
	raw, err := json.Marshal(entity)
	if err != nil {
		return -1, errors.Wrap(err, "Unmarshal entity")
	}

	n, err := x.w.Write(raw)
	if err != nil {
		if err == io.EOF {
			return 0, err
		}
		return -1, errors.Wrap(err, "Writing entity").With("entity", entity)
	}
	return n, nil
}

func NewEntityWriter(w io.Writer) EntityWriter {
	return &entityWriterImpl{
		w: w,
	}
}
