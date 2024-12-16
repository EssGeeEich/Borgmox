package Job

import (
	"net/http"
	"strings"
)

type JobData struct {
	BackupJobs map[string]BackupJobSettings
}

func (s *JobData) sendSuccessNotification(jobSettings BackupJobSettings, title string, message string, tags []string) error {
	if jobSettings.Notification.SuccessPriority == NP_Disabled {
		return nil
	}

	if req, err := http.NewRequest("POST", jobSettings.Notification.TargetServer+"/"+jobSettings.Notification.Topic,
		strings.NewReader(message)); err != nil {
		return err
	} else {
		req.Header.Set("Title", title)
		req.Header.Set("Priority", string(jobSettings.Notification.SuccessPriority))
		req.Header.Set("Tags", strings.Join(tags, ","))

		if res, err := http.DefaultClient.Do(req); err != nil {
			return err
		} else {
			defer res.Body.Close()
			// Post-process res?
			return nil
		}
	}
}

func (s *JobData) sendFailureNotification(jobSettings BackupJobSettings, title string, message string, tags []string) error {
	if jobSettings.Notification.FailurePriority == NP_Disabled {
		return nil
	}

	if req, err := http.NewRequest("POST", jobSettings.Notification.TargetServer+"/"+jobSettings.Notification.Topic,
		strings.NewReader(message)); err != nil {
		return err
	} else {
		req.Header.Set("Title", title)
		req.Header.Set("Priority", string(jobSettings.Notification.FailurePriority))
		req.Header.Set("Tags", strings.Join(tags, ","))

		if res, err := http.DefaultClient.Do(req); err != nil {
			return err
		} else {
			defer res.Body.Close()
			// Post-process res?
			return nil
		}
	}
}
