package Job

import (
	"encoding/base64"
	"net/http"
	"strings"
)

type JobData struct {
	BackupJobs map[string]BackupJobSettings
}

func highestPriority(a, b NotificationPriority) NotificationPriority {
	allNotificationTypes := []NotificationPriority{
		NP_Max,
		NP_Urgent,
		NP_High,
		NP_Default,
		NP_Low,
		NP_Min,
		NP_Disabled,
	}
	for _, v := range allNotificationTypes {
		if a == v || b == v {
			return v
		}
	}
	return NP_Disabled
}

func (s *JobData) sendSuccessNotification(jobSettings BackupJobSettings, notificationTargetInfo NotificationTargetInfo, title string, message string, tags []string) error {
	if notificationTargetInfo.SuccessPriority == NP_Disabled || jobSettings.Notification.TargetServer == "" {
		return nil
	}

	if req, err := http.NewRequest("POST", jobSettings.Notification.TargetServer+"/"+jobSettings.Notification.Topic,
		strings.NewReader(message)); err != nil {
		return err
	} else {
		if jobSettings.Notification.AuthUser != "" {
			req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(jobSettings.Notification.AuthUser+":"+jobSettings.Notification.AuthPassword)))
		} else if jobSettings.Notification.AuthPassword != "" {
			req.Header.Set("Authorization", "Bearer "+jobSettings.Notification.AuthPassword)
		}

		req.Header.Set("Title", title)
		req.Header.Set("Priority", string(notificationTargetInfo.SuccessPriority))
		req.Header.Set("Tags", strings.Join(tags, ","))

		if notificationTargetInfo.SuccessEmailTarget != "" {
			req.Header.Set("Email", notificationTargetInfo.SuccessEmailTarget)
		}

		if res, err := http.DefaultClient.Do(req); err != nil {
			return err
		} else {
			defer res.Body.Close()
			// Post-process res?
			return nil
		}
	}
}

func (s *JobData) sendFailureNotification(jobSettings BackupJobSettings, notificationTargetInfo NotificationTargetInfo, title string, message string, tags []string) error {
	if notificationTargetInfo.FailurePriority == NP_Disabled || jobSettings.Notification.TargetServer == "" {
		return nil
	}

	if req, err := http.NewRequest("POST", jobSettings.Notification.TargetServer+"/"+jobSettings.Notification.Topic,
		strings.NewReader(message)); err != nil {
		return err
	} else {
		if jobSettings.Notification.AuthUser != "" {
			req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(jobSettings.Notification.AuthUser+":"+jobSettings.Notification.AuthPassword)))
		} else if jobSettings.Notification.AuthPassword != "" {
			req.Header.Set("Authorization", "Bearer "+jobSettings.Notification.AuthPassword)
		}

		req.Header.Set("Title", title)
		req.Header.Set("Priority", string(notificationTargetInfo.FailurePriority))
		req.Header.Set("Tags", strings.Join(tags, ","))

		if notificationTargetInfo.FailureEmailTarget != "" {
			req.Header.Set("Email", notificationTargetInfo.FailureEmailTarget)
		}

		if res, err := http.DefaultClient.Do(req); err != nil {
			return err
		} else {
			defer res.Body.Close()
			// Post-process res?
			return nil
		}
	}
}
