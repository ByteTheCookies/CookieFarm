package api

import (
	"fmt"
	"logger"
	"models"
	"net/http"
	"net/url"
	"server/config"
	"server/database"
	"sharedconfig"

	json "github.com/bytedance/sonic"
)

func doJSON[T any](respBody []byte, out *T) error {
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("json decode error: %w", err)
	}
	return nil
}

func checkStatus(code int, body []byte) error {
	if code != 200 {
		return fmt.Errorf("status %d: %s", code, body)
	}
	return nil
}

func GetConfig() (sharedconfig.Shared, error) {
	client := getClient()
	cfg := config.GetInstance()

	logger.Log.Debug().Str("token", cfg.GetToken()).Msg("TOKEN")

	resp, body, err := client.get("/api/v1/config", AUTHED)
	if err != nil {
		return sharedconfig.Shared{}, err
	}
	defer resp.Body.Close()

	if err := checkStatus(resp.StatusCode, body); err != nil {
		logger.Log.Error().Msg(err.Error())
		return sharedconfig.Shared{}, err
	}

	var shcfg sharedconfig.Shared
	if err := doJSON(body, &shcfg); err != nil {
		return sharedconfig.Shared{}, err
	}

	return shcfg, nil
}

func Login(username, password string) error {
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)

	client := getClient()
	resp, body, err := client.postForm("/api/v1/auth/login", data, NOTAUTHED)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := checkStatus(resp.StatusCode, body); err != nil {
		return err
	}

	token, err := getCookie(resp, "token")
	if err != nil {
		return err
	}

	client.setToken(token)
	return nil
}

// @IMPORTANT: prefer ckp
func SubmitBatchDirect(flags []database.Flag) error {
	payload, err := json.Marshal(models.SubmitFlagsRequest{
		Flags: flags,
	})
	if err != nil {
		return err
	}

	client := getClient()
	resp, body, err := client.postJSON("/api/v1/submit-flags-standalone", payload, AUTHED)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return checkStatus(resp.StatusCode, body)
}

func SubmitFlag(flag database.Flag) error {
	payload, err := json.Marshal(models.SubmitFlagRequest{
		Flag: flag,
	})
	if err != nil {
		return err
	}

	client := getClient()
	resp, body, err := client.postJSON("/api/v1/submit-flag", payload, AUTHED)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return checkStatus(resp.StatusCode, body)
}

func UploadExploit(exploitPath string) error {
	client := getClient()
	logger.Log.Debug().Str("exploitPath", exploitPath).Msg("Uploading exploit")
	resp, body, err := client.uploadFile("/api/v1/exploit/upload", exploitPath, AUTHED)
	if err != nil {
		if resp.StatusCode == http.StatusConflict {
			return nil
		}
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		logger.Log.Debug().Str("exploitPath", exploitPath).Msg("Exploit already exists, skipping upload")
		return checkStatus(200, body)
	}

	return checkStatus(resp.StatusCode, body)
}
