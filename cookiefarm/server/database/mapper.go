package database

import (
	"protocols"
	"time"
)

func MapFromFlagToDBParams(params Flag) AddFlagParams {
	return AddFlagParams{
		FlagCode:     params.FlagCode,
		ServiceName:  params.ServiceName,
		TeamID:       params.TeamID,
		Username:     params.Username,
		ExploitName:  params.ExploitName,
		Status:       params.Status,
		Msg:          params.Msg,
		ResponseTime: params.ResponseTime,
		SubmitTime:   params.SubmitTime,
		PortService:  params.PortService,
	}
}

func MapFromResponseProtocolToParamsToUpdate(params protocols.ResponseProtocol) UpdateFlagStatusByCodeParams {
	return UpdateFlagStatusByCodeParams{
		FlagCode:     params.Flag,
		Status:       params.Status,
		Msg:          params.Msg,
		ResponseTime: uint64(time.Now().Unix()),
	}
}
