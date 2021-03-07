package main

import (
	"github.com/Cretezy/dSock/common"
)

func resolveClaims(options common.ResolveOptions, requestId string) ([]string, *common.ApiError) {
	if options.Session != "" {
		userSessionClaims := redisClient.SMembers("claim-user-session:" + options.User + "-" + options.Session)

		if userSessionClaims.Err() != nil {
			return nil, &common.ApiError{
				InternalError: userSessionClaims.Err(),
				ErrorCode:     common.ErrorGettingClaim,
				StatusCode:    500,
				RequestId:     requestId,
			}
		}

		return userSessionClaims.Val(), nil
	} else if options.Channel != "" {
		channelClaims := redisClient.SMembers("claim-channel:" + options.Channel)

		if channelClaims.Err() != nil {
			return nil, &common.ApiError{
				InternalError: channelClaims.Err(),
				ErrorCode:     common.ErrorGettingClaim,
				StatusCode:    500,
				RequestId:     requestId,
			}
		}

		return channelClaims.Val(), nil
	} else if options.User != "" {
		userClaims := redisClient.SMembers("claim-user:" + options.User)

		if userClaims.Err() != nil {
			return nil, &common.ApiError{
				InternalError: userClaims.Err(),
				ErrorCode:     common.ErrorGettingClaim,
				StatusCode:    500,
				RequestId:     requestId,
			}
		}

		return userClaims.Val(), nil
	} else {
		return []string{}, nil
	}
}
