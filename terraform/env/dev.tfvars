# Non-secret defaults for the dev environment.
# Only the Zoho SMTP password (zoho_pwd) is a real secret and is passed via
# -var in CI from the ZOHO_PWD GitHub secret — it must NOT be committed here.

s3_bucket_name = "prospect-mailer-app-data"
ecr_repo_name  = "prospect-mailer"

lambda_timeout = 120
lambda_memory  = 256

db_s3_key  = "mydata.db"
csv_s3_key = "prospects.csv"

# Non-secret SMTP settings.
sender    = "modernization@steadypartner.online"
smtp_host = "smtp.zoho.com"
smtp_port = "465"

resend_api_url = "https://api.resend.com/emails"
new_sender     = "modernization@steadypartner.co"