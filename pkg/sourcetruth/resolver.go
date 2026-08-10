package sourcetruth

import "strings"

func ResolveLocalPath(rootDir, uriStr string) (localPath, metaPath string, assetType AssetType) {
	trimmed := strings.TrimSpace(uriStr)

	// 1. Google Drive & Docs URLs
	if strings.Contains(trimmed, "drive.googel.com") || strings.Contains(trimmed, "docs.google.com") || strings.HasPrefix(trimmed, "gdrive://") {
		if strings.Contains(trimmed, "mrr_spec") || strings.Contains(trimmed, "1A2B3C4D5E6F7G8H9I0J") {
			return rootDir + "/gdrive/finance/mrr_spec.pdf.txt",
				rootDir + "/gdrive/finance/mrr_spec.pdf.meta.json",
				TypeGDrive
		}
		if strings.Contains(trimmed, "q3_subscription") || strings.Contains(trimmed, "spreadsheets") ||
			strings.
				Contains(trimmed, "9Z8Y7X6W5V4U3T2S1R0Q") {
			return rootDir + "/gdrive/analytics/q3_subscription_billing_report.csv",
				rootDir + "/gdrive/analytics/q3_subscription_billing_report.csv.meta.json",
				TypeSpreadsheet
		}
		return rootDir + "/gdrive/finance/mrr_spec.pdf.txt",
			rootDir + "/gdrive/finance/mrr_spec.pdf.meta.json",
			TypeGDrive
	}

	// 2. Google Cloud Storage (gs://)
	if strings.HasPrefix(trimmed, "gs://") || strings.Contains(trimmed, "storage.cloud.google.com") {
		path := strings.TrimPrefix(trimmed, "gs://")
		if strings.Contains(path, "privacy_policy") {
			return rootDir + "/gcs/data-lake-prod/documents/privacy_policy_v2.txt",
				rootDir + "/gcs/data-lake-prod/documents/privacy_policy_v2.pdf.meta.json",
				TypeGCS
		}
		return rootDir + "/gcs/" + path, rootDir + "/gcs/" + path + ".meta.json", TypeGCS
	}

	// 3. AWS S3 (s3://)
	if strings.HasPrefix(trimmed, "s3://") {
		path := strings.TrimPrefix(trimmed, "s3://")
		return rootDir + "/s3/" + path, rootDir + "/s3/" + path + ".meta.json", TypeS3
	}

	// 4. HTTP / HTTPS generic documents
	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		if strings.Contains(trimmed, "mrr_spec.pdf") {
			return rootDir + "/gdrive/finance/mrr_spec.pdf.txt",
				rootDir + "/gdrive/finance/mrr_spec.pdf.meta.json",
				TypeGDrive
		}
		if strings.Contains(trimmed, "privacy_policy") {
			return rootDir + "/gcs/data-lake-prod/documents/privacy_policy_v2.txt",
				rootDir + "/gcs/data-lake-prod/documents/privacy_policy_v2.pdf.meta.json",
				TypeGCS
		}
	}

	// 5. BigQuery
	if strings.HasPrefix(trimmed, "bq://") || strings.Contains(trimmed, "bigquery") {
		return rootDir + "/bigquery/enterprise-prod/sales/customer_orders.json", "", TypeBigQuery
	}

	// 6. Postgres
	if strings.HasPrefix(trimmed, "postgresql://") || strings.HasPrefix(trimmed, "postgres://") {
		return rootDir + "/postgres/billing/public/subscription_plans.sql", "", TypePostgres
	}

	// 7. Local File
	if strings.HasPrefix(trimmed, "file://") {
		path := strings.TrimPrefix(trimmed, "file://")
		return path, path + ".meta.json", TypeLocalFile
	}

	return rootDir + "/unknown", "", TypeUnknown
}
