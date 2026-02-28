package database

import (
	"protocols"
	"testing"
	"time"
)

// ─── MapFromFlagToDBParams ────────────────────────────────────────────────────

func TestMapFromFlagToDBParams_FullyPopulated_AllFieldsCopied(t *testing.T) {
	input := Flag{
		FlagCode:     "FLAG{mapper_001}",
		ServiceName:  "my_service",
		PortService:  9090,
		SubmitTime:   1_700_000_000,
		ResponseTime: 1_700_000_100,
		Msg:          "flag accepted",
		Status:       "ACCEPTED",
		TeamID:       42,
		Username:     "alice",
		ExploitName:  "exploit_alpha",
	}

	got := MapFromFlagToDBParams(input)

	if got.FlagCode != input.FlagCode {
		t.Errorf("FlagCode: want %q, got %q", input.FlagCode, got.FlagCode)
	}
	if got.ServiceName != input.ServiceName {
		t.Errorf("ServiceName: want %q, got %q", input.ServiceName, got.ServiceName)
	}
	if got.PortService != input.PortService {
		t.Errorf("PortService: want %d, got %d", input.PortService, got.PortService)
	}
	if got.SubmitTime != input.SubmitTime {
		t.Errorf("SubmitTime: want %d, got %d", input.SubmitTime, got.SubmitTime)
	}
	if got.ResponseTime != input.ResponseTime {
		t.Errorf("ResponseTime: want %d, got %d", input.ResponseTime, got.ResponseTime)
	}
	if got.Msg != input.Msg {
		t.Errorf("Msg: want %q, got %q", input.Msg, got.Msg)
	}
	if got.Status != input.Status {
		t.Errorf("Status: want %q, got %q", input.Status, got.Status)
	}
	if got.TeamID != input.TeamID {
		t.Errorf("TeamID: want %d, got %d", input.TeamID, got.TeamID)
	}
	if got.Username != input.Username {
		t.Errorf("Username: want %q, got %q", input.Username, got.Username)
	}
	if got.ExploitName != input.ExploitName {
		t.Errorf("ExploitName: want %q, got %q", input.ExploitName, got.ExploitName)
	}
}

func TestMapFromFlagToDBParams_ZeroValue_AllFieldsZero(t *testing.T) {
	input := Flag{}

	got := MapFromFlagToDBParams(input)

	if got.FlagCode != "" {
		t.Errorf("FlagCode: want empty string, got %q", got.FlagCode)
	}
	if got.ServiceName != "" {
		t.Errorf("ServiceName: want empty string, got %q", got.ServiceName)
	}
	if got.PortService != 0 {
		t.Errorf("PortService: want 0, got %d", got.PortService)
	}
	if got.SubmitTime != 0 {
		t.Errorf("SubmitTime: want 0, got %d", got.SubmitTime)
	}
	if got.ResponseTime != 0 {
		t.Errorf("ResponseTime: want 0, got %d", got.ResponseTime)
	}
	if got.Msg != "" {
		t.Errorf("Msg: want empty string, got %q", got.Msg)
	}
	if got.Status != "" {
		t.Errorf("Status: want empty string, got %q", got.Status)
	}
	if got.TeamID != 0 {
		t.Errorf("TeamID: want 0, got %d", got.TeamID)
	}
	if got.Username != "" {
		t.Errorf("Username: want empty string, got %q", got.Username)
	}
	if got.ExploitName != "" {
		t.Errorf("ExploitName: want empty string, got %q", got.ExploitName)
	}
}

func TestMapFromFlagToDBParams_MaxUint16Port_Preserved(t *testing.T) {
	input := sampleFlag("FLAG{mapper_maxport}")
	input.PortService = 65535 // max uint16

	got := MapFromFlagToDBParams(input)

	if got.PortService != 65535 {
		t.Errorf("PortService: want 65535, got %d", got.PortService)
	}
}

func TestMapFromFlagToDBParams_MaxUint16TeamID_Preserved(t *testing.T) {
	input := sampleFlag("FLAG{mapper_maxteam}")
	input.TeamID = 65535 // max uint16

	got := MapFromFlagToDBParams(input)

	if got.TeamID != 65535 {
		t.Errorf("TeamID: want 65535, got %d", got.TeamID)
	}
}

func TestMapFromFlagToDBParams_LargeTimestamps_Preserved(t *testing.T) {
	input := sampleFlag("FLAG{mapper_timestamps}")
	input.SubmitTime = 9_999_999_999
	input.ResponseTime = 9_999_999_998

	got := MapFromFlagToDBParams(input)

	if got.SubmitTime != 9_999_999_999 {
		t.Errorf("SubmitTime: want 9999999999, got %d", got.SubmitTime)
	}
	if got.ResponseTime != 9_999_999_998 {
		t.Errorf("ResponseTime: want 9999999998, got %d", got.ResponseTime)
	}
}

func TestMapFromFlagToDBParams_SpecialCharactersInStrings_Preserved(t *testing.T) {
	input := sampleFlag("FLAG{special_chars_!@#$%^&*()}")
	input.ServiceName = "service/with/slashes"
	input.Msg = "message with 'single quotes' and \"double quotes\""
	input.Username = "user@host.example.com"
	input.ExploitName = "exploit-v2_final.sh"

	got := MapFromFlagToDBParams(input)

	if got.FlagCode != input.FlagCode {
		t.Errorf("FlagCode with special chars not preserved: got %q", got.FlagCode)
	}
	if got.ServiceName != input.ServiceName {
		t.Errorf("ServiceName with slashes not preserved: got %q", got.ServiceName)
	}
	if got.Msg != input.Msg {
		t.Errorf("Msg with quotes not preserved: got %q", got.Msg)
	}
	if got.Username != input.Username {
		t.Errorf("Username not preserved: got %q", got.Username)
	}
	if got.ExploitName != input.ExploitName {
		t.Errorf("ExploitName not preserved: got %q", got.ExploitName)
	}
}

// Verify that the mapper returns a value type (not a pointer), so modifications
// to the result do not affect the original Flag.
func TestMapFromFlagToDBParams_ReturnsValueNotReference_OriginalUnchanged(t *testing.T) {
	input := sampleFlag("FLAG{mapper_copy}")

	got := MapFromFlagToDBParams(input)
	got.FlagCode = "FLAG{mutated}"
	got.Status = "MUTATED"

	// The original input must not have changed.
	if input.FlagCode == "FLAG{mutated}" {
		t.Error("MapFromFlagToDBParams returned a reference: mutating result changed input.FlagCode")
	}
	if input.Status == "MUTATED" {
		t.Error("MapFromFlagToDBParams returned a reference: mutating result changed input.Status")
	}
}

// ─── MapFromResponseProtocolToParamsToUpdate ──────────────────────────────────

func TestMapFromResponseProtocolToParamsToUpdate_FullyPopulated_FieldsMapped(t *testing.T) {
	input := protocols.ResponseProtocol{
		Flag:   "FLAG{response_001}",
		Status: "ACCEPTED",
		Msg:    "congratulations",
	}

	got := MapFromResponseProtocolToParamsToUpdate(input)

	if got.FlagCode != input.Flag {
		t.Errorf("FlagCode: want %q, got %q", input.Flag, got.FlagCode)
	}
	if got.Status != input.Status {
		t.Errorf("Status: want %q, got %q", input.Status, got.Status)
	}
	if got.Msg != input.Msg {
		t.Errorf("Msg: want %q, got %q", input.Msg, got.Msg)
	}
}

func TestMapFromResponseProtocolToParamsToUpdate_ZeroValue_StringFieldsEmptyTimestampSet(t *testing.T) {
	before := uint64(time.Now().Unix())
	input := protocols.ResponseProtocol{}

	got := MapFromResponseProtocolToParamsToUpdate(input)

	after := uint64(time.Now().Unix())

	if got.FlagCode != "" {
		t.Errorf("FlagCode: want empty string, got %q", got.FlagCode)
	}
	if got.Status != "" {
		t.Errorf("Status: want empty string, got %q", got.Status)
	}
	if got.Msg != "" {
		t.Errorf("Msg: want empty string, got %q", got.Msg)
	}
	// Even for a zero-value input, ResponseTime must be a current timestamp.
	if got.ResponseTime < before || got.ResponseTime > after+1 {
		t.Errorf(
			"ResponseTime should be a current Unix timestamp in [%d, %d], got %d",
			before, after+1, got.ResponseTime,
		)
	}
}

// Issue 4.5 — FIXED: ResponseTime is now set to time.Now().Unix() by the mapper.
// This test verifies that the returned ResponseTime is a plausible current
// Unix timestamp (within a ±5 second window of the call to allow for slow CI).
func TestMapFromResponseProtocolToParamsToUpdate_ResponseTimeIsCurrentTimestamp(t *testing.T) {
	before := uint64(time.Now().Unix())

	input := protocols.ResponseProtocol{
		Flag:   "FLAG{rt_timestamp}",
		Status: "DENIED",
		Msg:    "already submitted",
	}

	got := MapFromResponseProtocolToParamsToUpdate(input)

	after := uint64(time.Now().Unix())

	if got.ResponseTime < before || got.ResponseTime > after+1 {
		t.Errorf(
			"ResponseTime should be a current Unix timestamp in [%d, %d], got %d",
			before, after+1, got.ResponseTime,
		)
	}
}

func TestMapFromResponseProtocolToParamsToUpdate_StatusVariants_AllPreserved(t *testing.T) {
	statuses := []string{"ACCEPTED", "DENIED", "ERROR", "UNSUBMITTED", "custom_status"}

	for _, status := range statuses {
		t.Run("status_"+status, func(t *testing.T) {
			input := protocols.ResponseProtocol{
				Flag:   "FLAG{status_test}",
				Status: status,
				Msg:    "msg for " + status,
			}

			got := MapFromResponseProtocolToParamsToUpdate(input)

			if got.Status != status {
				t.Errorf("Status: want %q, got %q", status, got.Status)
			}
		})
	}
}

func TestMapFromResponseProtocolToParamsToUpdate_FlagFieldMapped_NotMsg(t *testing.T) {
	// Regression guard: the mapper must copy ResponseProtocol.Flag into
	// UpdateFlagStatusByCodeParams.FlagCode, not into Msg or any other field.
	input := protocols.ResponseProtocol{
		Flag:   "FLAG{field_guard}",
		Status: "ACCEPTED",
		Msg:    "different_value",
	}

	got := MapFromResponseProtocolToParamsToUpdate(input)

	if got.FlagCode != "FLAG{field_guard}" {
		t.Errorf("FlagCode should come from ResponseProtocol.Flag; got %q", got.FlagCode)
	}
	if got.Msg == "FLAG{field_guard}" {
		t.Error("Msg must not be populated with the Flag value — field mapping is crossed")
	}
}

func TestMapFromResponseProtocolToParamsToUpdate_ReturnsValueNotReference(t *testing.T) {
	input := protocols.ResponseProtocol{
		Flag:   "FLAG{ref_guard}",
		Status: "ACCEPTED",
		Msg:    "ok",
	}

	got := MapFromResponseProtocolToParamsToUpdate(input)
	got.FlagCode = "FLAG{mutated}"

	if input.Flag == "FLAG{mutated}" {
		t.Error("MapFromResponseProtocolToParamsToUpdate returned a reference: mutating result changed input")
	}
}
