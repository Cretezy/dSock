package main

import "github.com/Cretezy/dSock/common"

func resolveClaims(options common.ResolveOptions) ([]string, *common.ApiError) {
	if options.Session != "" {
		userSessionClaims := redisClient.SMembers("claim-user-session:" + options.User + "-" + options.Session)

		if userSessionClaims.Err() != nil {
			return nil, &common.ApiError{
				ErrorCode:  common.ErrorGettingClaim,
				StatusCode: 500,
			}
		}

		return userSessionClaims.Val(), nil
	} else if options.User != "" {
		userClaims := redisClient.SMembers("claim-user:" + options.User)

		if userClaims.Err() != nil {
			return nil, &common.ApiError{
				ErrorCode:  common.ErrorGettingClaim,
				StatusCode: 500,
			}
		}

		return userClaims.Val(), nil
	} else {
		return []string{}, nil
	}
}
