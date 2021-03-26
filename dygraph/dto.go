package dygraph

import (
	"github.com/google/uuid"
	"strings"
)

type dto struct {
	GlobalId       string `json:"globalId"`
	TypeTarget     string `json:"typeTarget"`
	OrganisationId string `json:"organisationId"`
	Id             string `json:"id"`
	Type           string `json:"type"`
	Data           string `json:"data"`
}

func (n dto) createLogicalRecord() LogicalRecord {
	return LogicalRecord{
		LogicalRecordRequest: LogicalRecordRequest{
			OrganisationId: uuid.MustParse(n.OrganisationId),
			Id:             uuid.MustParse(n.Id),
			Type:           n.Type,
			Data:           n.Data,
		},
		TypeTarget: strings.Split(n.TypeTarget, separator),
	}
}
