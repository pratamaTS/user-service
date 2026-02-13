package dao

type SubscriptionType string

const (
	SubscriptionTrial    SubscriptionType = "trial"
	SubscriptionBusiness SubscriptionType = "business"
	SubscriptionCompany  SubscriptionType = "company"
)
