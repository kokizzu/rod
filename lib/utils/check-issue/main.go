// The .github/workflows/check-issues.yml will use it as an github action
// To test it locally, you can generate a personal github token: https://github.com/settings/tokens
// Then run this:
//   ROD_GITHUB_ROBOT=your_token go run ./lib/utils/check-issue

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/go-rod/rod/lib/utils"
	"github.com/ysmood/gson"
)

var token = os.Getenv("ROD_GITHUB_ROBOT")
var issuePath = os.Getenv("GITHUB_EVENT_PATH")

const ref = "\n_<sub>generated by [check-issue](https://github.com/go-rod/rod/tree/master/lib/utils/check-issue)</sub>_"

func main() {
	if issuePath == "" {
		issuePath = "lib/utils/check-issue/issue.json"
	}

	data, err := os.Open(issuePath)
	utils.E(err)

	issue := gson.New(data).Get("issue")

	for _, l := range issue.Get("labels").Arr() {
		name := l.Get("name").Str()
		if name != "question" && name != "bug" {
			log.Println("skip", name)
			return
		}
	}

	num := issue.Get("number").Int()
	body := issue.Get("body").Str()

	log.Println("check issue", num)

	m := regexp.MustCompile(`\*\*Rod Version:\*\* v[0-9.]+`).FindString(body)
	if m == "" || m == "**Rod Version:** v0.0.0" {
		log.Println("invalid issue format")

		msg := fmt.Sprintf(
			"Please add a valid `**Rod Version:** v0.0.0` to your issue. Current version is %s"+
				ref,
			currentVer(),
		)

		q := req(fmt.Sprintf("/repos/go-rod/rod/issues/%d/comments", num))
		q.Method = http.MethodPost
		q.Body = ioutil.NopCloser(bytes.NewBuffer(utils.MustToJSONBytes(map[string]string{"body": msg})))
		res, err := http.DefaultClient.Do(q)
		utils.E(err)

		log.Println(res.Status)

		if res.StatusCode >= 400 {
			str, err := ioutil.ReadAll(res.Body)
			utils.E(err)
			log.Fatal(string(str))
		}
	}
}

func currentVer() string {
	q := req("/repos/go-rod/rod/tags?per_page=1")
	res, err := http.DefaultClient.Do(q)
	utils.E(err)

	currentVer := gson.New(res.Body).Get("0.name").Str()

	return currentVer
}

func req(u string) *http.Request {
	r, err := http.NewRequest(http.MethodGet, "https://api.github.com"+u, nil)
	utils.E(err)
	r.Header.Add("Authorization", "token "+token)
	return r
}
