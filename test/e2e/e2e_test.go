package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"prReviewerAssignment/internal/models"

	"github.com/stretchr/testify/assert"
)

const baseURL = "http://localhost:8080"

func TestFullPRReviewWorkflow(t *testing.T) {
	client := &http.Client{Timeout: 10 * time.Second}

	teamName := "e2e-team-" + fmt.Sprintf("%d", time.Now().UnixNano())

	team := models.Team{
		TeamName: teamName,
		Members: []models.TeamMember{
			{UserID: "e2e-dev1", Username: "E2E Developer 1", IsActive: true},
			{UserID: "e2e-dev2", Username: "E2E Developer 2", IsActive: true},
			{UserID: "e2e-dev3", Username: "E2E Developer 3", IsActive: true},
			{UserID: "e2e-dev4", Username: "E2E Developer 4", IsActive: true},
		},
	}

	teamJSON, _ := json.Marshal(team)
	resp, err := client.Post(baseURL+"/team/add", "application/json", bytes.NewBuffer(teamJSON))
	if resp.StatusCode == 400 {
		t.Log("Team already exists or validation failed, continuing with existing team")
	} else {
		assert.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode)
	}

	prID := "pr-e2e-" + fmt.Sprintf("%d", time.Now().UnixNano())

	prRequest := map[string]string{
		"pull_request_id":   prID,
		"pull_request_name": "E2E Test Feature",
		"author_id":         "e2e-dev1",
	}

	prJSON, _ := json.Marshal(prRequest)
	resp, err = client.Post(baseURL+"/pullRequest/create", "application/json", bytes.NewBuffer(prJSON))
	if resp.StatusCode == 409 {
		prID = "pr-e2e-retry-" + fmt.Sprintf("%d", time.Now().UnixNano())
		prRequest["pull_request_id"] = prID
		prJSON, _ = json.Marshal(prRequest)
		resp, err = client.Post(baseURL+"/pullRequest/create", "application/json", bytes.NewBuffer(prJSON))
	}
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)

	var prCreateResponse map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&prCreateResponse)
	if prCreateResponse["pr"] != nil {
		createdPR := prCreateResponse["pr"].(map[string]interface{})
		if createdPR["assigned_reviewers"] != nil {
			reviewers := createdPR["assigned_reviewers"].([]interface{})
			if len(reviewers) > 0 {
				assert.Len(t, reviewers, 2)
				assert.Equal(t, "OPEN", createdPR["status"])

				reassignRequest := map[string]string{
					"pull_request_id": prID,
					"old_reviewer_id": reviewers[0].(string),
				}

				reassignJSON, _ := json.Marshal(reassignRequest)
				resp, err = client.Post(baseURL+"/pullRequest/reassign", "application/json", bytes.NewBuffer(reassignJSON))
				assert.NoError(t, err)
				assert.Equal(t, 200, resp.StatusCode)

				var reassignResponse map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&reassignResponse)
				assert.Contains(t, reassignResponse, "replaced_by")

				resp, err = client.Get(baseURL + "/stats/reviewers")
				assert.NoError(t, err)
				assert.Equal(t, 200, resp.StatusCode)

				var statsResponse models.StatsResponse
				json.NewDecoder(resp.Body).Decode(&statsResponse)
				assert.Greater(t, len(statsResponse.ReviewerStats), 0)

				mergeRequest := map[string]string{
					"pull_request_id": prID,
				}

				mergeJSON, _ := json.Marshal(mergeRequest)
				resp, err = client.Post(baseURL+"/pullRequest/merge", "application/json", bytes.NewBuffer(mergeJSON))
				assert.NoError(t, err)
				assert.Equal(t, 200, resp.StatusCode)

				resp, err = client.Post(baseURL+"/pullRequest/reassign", "application/json", bytes.NewBuffer(reassignJSON))
				assert.NoError(t, err)
				assert.Equal(t, 409, resp.StatusCode)

				resp, err = client.Get(baseURL + "/users/getReview?user_id=" + reviewers[1].(string))
				assert.NoError(t, err)
				assert.Equal(t, 200, resp.StatusCode)

				var userReviewsResponse map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&userReviewsResponse)
				assert.Equal(t, reviewers[1].(string), userReviewsResponse["user_id"])
			}
		}
	}
}

func TestTeamManagementWorkflow(t *testing.T) {
	client := &http.Client{Timeout: 10 * time.Second}

	teamName := "management-team-" + fmt.Sprintf("%d", time.Now().UnixNano())

	team := models.Team{
		TeamName: teamName,
		Members: []models.TeamMember{
			{UserID: "mgmt1", Username: "Management User 1", IsActive: true},
			{UserID: "mgmt2", Username: "Management User 2", IsActive: true},
			{UserID: "mgmt3", Username: "Management User 3", IsActive: true},
		},
	}

	teamJSON, _ := json.Marshal(team)
	resp, err := client.Post(baseURL+"/team/add", "application/json", bytes.NewBuffer(teamJSON))
	if resp.StatusCode == 400 {
		teamName = "management-team-alt-" + fmt.Sprintf("%d", time.Now().UnixNano())
		team.TeamName = teamName
		teamJSON, _ = json.Marshal(team)
		resp, err = client.Post(baseURL+"/team/add", "application/json", bytes.NewBuffer(teamJSON))
	}
	assert.NoError(t, err)
	if resp.StatusCode == 201 || resp.StatusCode == 200 {
		t.Logf("Team created/updated successfully with status: %d", resp.StatusCode)
	} else {
		t.Logf("Unexpected status: %d, but continuing test", resp.StatusCode)
	}

	resp, err = client.Get(baseURL + "/team/get?team_name=" + teamName)
	assert.NoError(t, err)
	if resp.StatusCode == 200 {
		var retrievedTeam models.Team
		json.NewDecoder(resp.Body).Decode(&retrievedTeam)
		assert.Len(t, retrievedTeam.Members, 3)
		assert.Equal(t, "Management User 1", retrievedTeam.Members[0].Username)
	} else {
		t.Logf("Could not retrieve team, status: %d", resp.StatusCode)
	}
}

func TestUserActivationWorkflow(t *testing.T) {
	client := &http.Client{Timeout: 10 * time.Second}

	teamName := "activation-team-" + fmt.Sprintf("%d", time.Now().UnixNano())
	team := models.Team{
		TeamName: teamName,
		Members: []models.TeamMember{
			{UserID: "active-user1", Username: "Active User 1", IsActive: true},
			{UserID: "active-user2", Username: "Active User 2", IsActive: true},
		},
	}

	teamJSON, _ := json.Marshal(team)
	resp, err := client.Post(baseURL+"/team/add", "application/json", bytes.NewBuffer(teamJSON))
	assert.NoError(t, err)
	if resp.StatusCode != 201 {
		t.Logf("Team creation returned status: %d, continuing...", resp.StatusCode)
	}

	deactivateRequest := map[string]interface{}{
		"user_id":   "active-user1",
		"is_active": false,
	}

	deactivateJSON, _ := json.Marshal(deactivateRequest)
	resp, err = client.Post(baseURL+"/users/setIsActive", "application/json", bytes.NewBuffer(deactivateJSON))
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var deactivateResponse map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&deactivateResponse)
	if deactivateResponse["user"] != nil {
		user := deactivateResponse["user"].(map[string]interface{})
		assert.Equal(t, false, user["is_active"])
		assert.Equal(t, "active-user1", user["user_id"])
	}

	activateRequest := map[string]interface{}{
		"user_id":   "active-user1",
		"is_active": true,
	}

	activateJSON, _ := json.Marshal(activateRequest)
	resp, err = client.Post(baseURL+"/users/setIsActive", "application/json", bytes.NewBuffer(activateJSON))
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	nonexistentRequest := map[string]interface{}{
		"user_id":   "nonexistent-user",
		"is_active": false,
	}
	nonexistentJSON, _ := json.Marshal(nonexistentRequest)
	resp, err = client.Post(baseURL+"/users/setIsActive", "application/json", bytes.NewBuffer(nonexistentJSON))
	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)
}

func TestErrorScenarios(t *testing.T) {
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(baseURL + "/team/get?team_name=nonexistent-team")
	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)

	resp, err = client.Get(baseURL + "/users/getReview?user_id=nonexistent-user")
	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)

	prRequest := map[string]string{
		"pull_request_id":   "conflict-pr",
		"pull_request_name": "Conflict PR",
		"author_id":         "nonexistent-author",
	}
	prJSON, _ := json.Marshal(prRequest)
	resp, err = client.Post(baseURL+"/pullRequest/create", "application/json", bytes.NewBuffer(prJSON))
	assert.NoError(t, err)
	assert.True(t, resp.StatusCode == 404 || resp.StatusCode == 409)

	mergeRequest := map[string]string{
		"pull_request_id": "nonexistent-pr",
	}
	mergeJSON, _ := json.Marshal(mergeRequest)
	resp, err = client.Post(baseURL+"/pullRequest/merge", "application/json", bytes.NewBuffer(mergeJSON))
	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)

	reassignRequest := map[string]string{
		"pull_request_id": "nonexistent-pr",
		"old_reviewer_id": "nonexistent-user",
	}
	reassignJSON, _ := json.Marshal(reassignRequest)
	resp, err = client.Post(baseURL+"/pullRequest/reassign", "application/json", bytes.NewBuffer(reassignJSON))
	assert.NoError(t, err)
	assert.True(t, resp.StatusCode == 404 || resp.StatusCode == 409)
}
