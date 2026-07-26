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

	// 7. Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if nil != err {
		slog.Error("Error sending request", slog.String("error msg", err.Error()))
		return false
	}
	defer resp.Body.Close()

	// 8. Print the response status code
	slog.Info("Status Code", "Code", resp.StatusCode)

	// 9. Read and print the response body
	var apiResponse map[string]interface{}

	if 200 == resp.StatusCode {
		err := json.NewDecoder(resp.Body).Decode(&apiResponse)
		if nil != err {
			slog.Error("Error parsing response", slog.String("error msg", err.Error()))
			return false
		}
		slog.Info("API Response", "message", apiResponse)
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
	subject := fmt.Sprintf(`Modernizing %s's backend`, company)

	msgBody := fmt.Sprintf(`%s,

Many engineering teams face bottlenecks where legacy debt slows down feature delivery and inflates Cloud costs.

I am a Senior Software Engineer specializing in end-to-end system modernization. I help teams and managers safely modernize complex legacy monoliths (especially into highly optimized cloud native Go or Python/FastAPI architectures), potentially reducing deployment friction and cloud spend by 40-60%%, while maintaining production stability and operational simplicity.

Because I handle the entire pipeline and lifecycle - from needs analysis and solution design to IaC (Terraform) and production monitoring - I operate as an autonomous execution unit without requiring daily management overhead. I also leverage safe AI-augmented workflows to compress timelines by 30-40%%.

I operate under an independent B2B contractor framework aligned with your working hours (using standard W-8BEN compliance for seamless US invoicing), meaning zero payroll or HR overhead for your company.

If you need an extra senior capacity to clear legacy bottlenecks or accelerate cloud migrations, let's connect.

Best regards,

Alexandre NGUYEN
Senior Software Engineer

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
