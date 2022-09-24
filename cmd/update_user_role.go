/*
 * Copyright (C) 2022 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */

package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"intel/amber/tac/v1/client/tms"
	"intel/amber/tac/v1/config"
	"intel/amber/tac/v1/constants"
	"intel/amber/tac/v1/models"
	"net/http"
	"net/url"
	"time"

	"github.com/spf13/cobra"
)

var updateUserRoleCmd = &cobra.Command{
	Use:   constants.RoleCmd,
	Short: "Updates role of a user under a tenant",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Info("update user role called")
		userId, err := updateUserRole(cmd)
		if err != nil {
			return err
		}
		fmt.Printf("\nUpdated User: %s \n\n", userId)
		return nil
	},
}

func init() {
	updateUserCmd.AddCommand(updateUserRoleCmd)

	updateUserRoleCmd.Flags().StringVarP(&apiKey, constants.ApiKeyParamName, "a", "", "API key to be used to connect to amber services")
	updateUserRoleCmd.Flags().StringP(constants.TenantIdParamName, "t", "", "Id of the tenant for whom the user needs to be created")
	updateUserRoleCmd.Flags().StringP(constants.UserIdParamName, "u", "", "Id of the specific user")
	updateUserRoleCmd.Flags().StringSliceP(constants.UserRoleParamName, "r", []string{}, "Comma separated roles of the specific user to be updated. Should be either Tenant Admin or User")
	updateUserRoleCmd.MarkFlagRequired(constants.ApiKeyParamName)
	updateUserRoleCmd.MarkFlagRequired(constants.UserIdParamName)
	updateUserRoleCmd.MarkFlagRequired(constants.UserRoleParamName)
}

func updateUserRole(cmd *cobra.Command) (string, error) {
	configValues, err := config.LoadConfiguration()
	if err != nil {
		return "", err
	}
	client := &http.Client{
		Timeout: time.Duration(configValues.HTTPClientTimeout) * time.Second,
	}

	tmsUrl, err := url.Parse(configValues.AmberBaseUrl + constants.TmsBaseUrl)
	if err != nil {
		return "", err
	}

	tenantIdString, err := cmd.Flags().GetString(constants.TenantIdParamName)
	if err != nil {
		return "", err
	}

	if tenantIdString == "" {
		tenantIdString = configValues.TenantId
	}

	tenantId, err := uuid.Parse(tenantIdString)
	if err != nil {
		return "", errors.Wrap(err, "Invalid tenant id provided")
	}

	userIdString, err := cmd.Flags().GetString(constants.UserIdParamName)
	if err != nil {
		return "", err
	}

	userId, err := uuid.Parse(userIdString)
	if err != nil {
		return "", errors.Wrap(err, "Invalid user id provided")
	}

	userRoles, err := cmd.Flags().GetStringSlice(constants.UserRoleParamName)
	if err != nil {
		return "", err
	}

	if len(userRoles) == 0 {
		return "", errors.New("User role cannot be empty")
	}

	for _, role := range userRoles {
		if role != constants.TenantAdminRole && role != constants.UserRole {
			return "", errors.Errorf("%s is not a valid user role. Roles should be either %s or %s", role,
				constants.TenantAdminRole, constants.UserRole)
		}
	}

	updateUserRoleReq := &models.UpdateTenantUserRoles{
		UserId: userId,
		Roles:  userRoles,
	}

	tmsClient := tms.NewTmsClient(client, tmsUrl, tenantId, apiKey)

	response, err := tmsClient.UpdateTenantUserRole(updateUserRoleReq)
	if err != nil {
		return "", err
	}

	responseBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return "", err
	}

	return string(responseBytes), nil
}
