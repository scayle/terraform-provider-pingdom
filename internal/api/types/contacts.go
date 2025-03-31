package api_types

type Contacts struct {
	// A list of all contacts in the organization
	Contacts []Contact `json:"contacts"`
}

type Contact struct {
	// Contact ID
	Id int64 `json:"id"`
	// Contact name
	Name string `json:"name"`
	// Describes whether alerts are paused for this contact
	Paused bool `json:"paused"`
	// Type defines whether this is a user (login user) or a contact only
	// One of: "user" or "contact"
	Type string `json:"type"`
	// Indicates whether the contact is the owner of the organization
	Owner               bool                `json:"owner"`
	NotificationTargets NotificationTargets `json:"notification_targets"`
}

type NotificationTargets struct {
	// A list of emails that will get notified for this contact
	Emails []EmailNotificationTarget `json:"email"`
}

type EmailNotificationTarget struct {
	// Contact target's severity level
	Severity string `json:"severity"`
	// Email address
	Address string `json:"address"`
}
