package scheduler

import "gopcpv2-web-spider/errors"

func genError(errMsg string) error {
	return errors.NewCrawlerError(errors.ERROR_TYPE_SCHEDULER, errMsg)
}