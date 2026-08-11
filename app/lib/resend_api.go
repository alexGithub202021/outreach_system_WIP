package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

type RequestBody struct {
	Text    string `json:"text"`
	To      string `json:"to"` // <--- THIS IS THE MISSING FIELD
	From    string `json:"from"`
	Subject string `json:"subject"`
}

type Response struct {
	StatusCode int    `json:"statusCode"`
	Name       string `json:"name"`
	Message    string `json:"message"`
}

func CallResendApi(to string, name string, company string) bool {

	reqBody := getRequestBody(to, company, name)

	// 5. Set up the HTTP POST request
	resendUrl := os.Getenv("RESEND_API_URL")

	req, err := http.NewRequest("POST", resendUrl, reqBody)
	if err != nil {
		panic(err)
	}

	// 6. Set essential headers
	api_key := fmt.Sprintf("Bearer %s", os.Getenv("RESEND_API_KEY"))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", api_key)
	req.Header.Set("List-Unsubscribe", "<mailto:unsubscribe@steadypartner.co?subject=unsubscribe>, <https://steadypartner.co/unsubscribe>")
	req.Header.Set("List-Unsubscribe-Post", "List-Unsubscribe=One-Click")

	// 7. Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if nil != err {
		slog.Error("Error sending request", slog.String("error msg", err.Error()))
		return false
	}
	defer resp.Body.Close()

	// 8. Print the response status code
	slog.Info("CallResendApi", "API Response Status Code", resp.StatusCode)

	// 9. Read and print the response body
	var apiResponse map[string]interface{}

	if 200 == resp.StatusCode {
		err := json.NewDecoder(resp.Body).Decode(&apiResponse)
		if nil != err {
			slog.Error("Error parsing response", slog.String("error msg", err.Error()))
			return false
		}
		slog.Info("CallResendApi", "Resend API Response", apiResponse)
		slog.Info("CallResendApi > email sent to:",
			"prospect name", name,
			"prospect company", company,
		)
	} else {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		slog.Error("API Error", slog.String("error msg", buf.String()))
		return false
	}

	return true
}

func getRequestBody(to string, company string, name string) *bytes.Buffer {

	// 1. Define the input data
	recipientTo := to // <--- YOU MUST DEFINE THIS VARIABLE
	sender := os.Getenv("NEW_SENDER")
	subject := fmt.Sprintf(`quick question re: Modernizing %s's backend`, company)

	msgBody := fmt.Sprintf(`%s,

Many engineering teams face bottlenecks where legacy debt slows down feature delivery and inflates cloud costs.

I'm a Senior Software Engineer specializing in end-to-end legacy modernization, refactoring monoliths into optimized cloud-native architectures, migrating to Go / FastAPI + Terraform when needed.
I help teams clear backend bottlenecks, reduce cloud spend, and compress delivery timelines without adding management overhead.

Available for an immediate start on a 3–6 month engagement. I operate as an autonomous B2B contractor, fully aligned with your working hours and with zero HR friction.

Do you have 10 minutes this week to discuss if I can help clear one of your current bottlenecks?

Best regards,

Alexandre NGUYEN
Senior Software Engineer & Systems Architect
(US W-8BEN compliant / EU timezone-aligned)

Email: modernization@steadypartner.co
GitHub: github.com/alexGithub202021
LinkedIn: https://www.linkedin.com/in/alexandre-nguyen-senior-swe`, name)

	// 2. Create the JSON payload including the new 'to' field
	jsonData := RequestBody{
		Text:    msgBody,
		To:      recipientTo, // <--- ADD THIS
		From:    sender,
		Subject: subject,
	}

	// 4. Convert the Go struct to a JSON byte slice
	jsonBytes, err := json.Marshal(jsonData)
	if err != nil {
		panic(err)
	}
	return bytes.NewBuffer(jsonBytes)
}
