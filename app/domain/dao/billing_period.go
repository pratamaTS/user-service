package dao

type BillingPeriod string

const (
	BillingWeekly  BillingPeriod = "weekly"
	BillingMonthly BillingPeriod = "monthly"
	BillingYearly  BillingPeriod = "yearly"
)
