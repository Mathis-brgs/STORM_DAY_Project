package message

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gateway/internal/models"

	apiv1 "github.com/Mathis-brgs/storm-project/services/message/api/v1"
	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/proto"
)

func (h *Handler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	var req models.CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.GroupResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: "invalid JSON"},
		})
		return
	}

	actorID := h.actorIDFromToken(r)
	if actorID == "" {
		respondJSON(w, http.StatusBadRequest, models.GroupResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: "actor_id (or user_id / X-User-ID) required"},
		})
		return
	}

	protoReq := &apiv1.GroupCreateRequest{
		ActorId:   actorID,
		Name:      req.Name,
		AvatarUrl: req.AvatarURL,
	}
	data, err := proto.Marshal(protoReq)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.GroupResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "INTERNAL", Message: err.Error()},
		})
		return
	}

	reply, err := h.nc.Request(subjectGroupCreate, data, requestTimeout)
	if err != nil {
		respondJSON(w, http.StatusBadGateway, models.GroupResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "GATEWAY_ERROR", Message: "message-service unreachable: " + err.Error()},
		})
		return
	}

	var resp apiv1.GroupCreateResponse
	if err := proto.Unmarshal(reply.Data, &resp); err != nil {
		respondJSON(w, http.StatusBadGateway, models.GroupResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "GATEWAY_ERROR", Message: "invalid response from message-service"},
		})
		return
	}

	out := models.GroupResponse{OK: resp.GetOk()}
	if resp.GetData() != nil {
		out.Data = toGroupModel(resp.GetData())
	}
	if resp.GetError() != nil {
		out.Error = &models.SendMessageError{
			Code:    resp.GetError().GetCode(),
			Message: resp.GetError().GetMessage(),
		}
	}

	// Ajouter les membres supplémentaires si member_ids est fourni
	if resp.GetOk() && resp.GetData() != nil && len(req.MemberIDs) > 0 {
		conversationID := resp.GetData().GetId()
		for _, memberID := range req.MemberIDs {
			if memberID == actorID {
				continue // le créateur est déjà membre (owner)
			}
			addReq := &apiv1.GroupAddMemberRequest{
				ActorId:        actorID,
				ConversationId: conversationID,
				GroupId:        conversationID,
				UserId:         memberID,
				Role:           0, // member
			}
			addData, err := proto.Marshal(addReq)
			if err != nil {
				log.Printf("[Gateway] CreateGroup: marshal add-member error: %v", err)
				continue
			}
			if _, err := h.nc.Request(subjectGroupAddMember, addData, requestTimeout); err != nil {
				log.Printf("[Gateway] CreateGroup: add-member %s error: %v", memberID, err)
			}
		}
	}

	status := http.StatusOK
	if !resp.GetOk() && resp.GetError() != nil {
		status = statusFromServiceCode(resp.GetError().GetCode(), http.StatusUnprocessableEntity)
	}
	respondJSON(w, status, out)
}

func (h *Handler) GetGroup(w http.ResponseWriter, r *http.Request) {
	conversationID, ok := groupIDFromPath(r)
	if !ok {
		respondJSON(w, http.StatusBadRequest, models.GroupResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: invalidId},
		})
		return
	}

	actorID := h.actorIDFromToken(r)
	if actorID == "" {
		respondJSON(w, http.StatusBadRequest, models.GroupResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: "actor_id (or user_id / X-User-ID) required"},
		})
		return
	}

	protoReq := &apiv1.GroupGetRequest{
		ActorId:        actorID,
		ConversationId: int32(conversationID),
		GroupId:        int32(conversationID),
	}
	data, err := proto.Marshal(protoReq)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.GroupResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "INTERNAL", Message: err.Error()},
		})
		return
	}

	reply, err := h.nc.Request(subjectGroupGet, data, requestTimeout)
	if err != nil {
		respondJSON(w, http.StatusBadGateway, models.GroupResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "GATEWAY_ERROR", Message: "message-service unreachable: " + err.Error()},
		})
		return
	}

	var resp apiv1.GroupGetResponse
	if err := proto.Unmarshal(reply.Data, &resp); err != nil {
		respondJSON(w, http.StatusBadGateway, models.GroupResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "GATEWAY_ERROR", Message: "invalid response from message-service"},
		})
		return
	}

	out := models.GroupResponse{OK: resp.GetOk()}
	if resp.GetData() != nil {
		out.Data = toGroupModel(resp.GetData())
		h.resolveGroupDisplayName(out.Data, actorID)
	}
	if resp.GetError() != nil {
		out.Error = &models.SendMessageError{
			Code:    resp.GetError().GetCode(),
			Message: resp.GetError().GetMessage(),
		}
	}

	status := http.StatusOK
	if !resp.GetOk() && resp.GetError() != nil {
		status = statusFromServiceCode(resp.GetError().GetCode(), http.StatusUnprocessableEntity)
	}
	respondJSON(w, status, out)
}

func (h *Handler) ListGroups(w http.ResponseWriter, r *http.Request) {
	actorID := h.actorIDFromToken(r)
	if actorID == "" {
		respondJSON(w, http.StatusBadRequest, models.GroupsResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: "actor_id (or user_id / X-User-ID) required"},
		})
		return
	}

	protoReq := &apiv1.GroupListForUserRequest{UserId: actorID}
	data, err := proto.Marshal(protoReq)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.GroupsResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "INTERNAL", Message: err.Error()},
		})
		return
	}

	reply, err := h.nc.Request(subjectGroupListForUser, data, requestTimeout)
	if err != nil {
		respondJSON(w, http.StatusBadGateway, models.GroupsResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "GATEWAY_ERROR", Message: "message-service unreachable: " + err.Error()},
		})
		return
	}

	var resp apiv1.GroupListForUserResponse
	if err := proto.Unmarshal(reply.Data, &resp); err != nil {
		respondJSON(w, http.StatusBadGateway, models.GroupsResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "GATEWAY_ERROR", Message: "invalid response from message-service"},
		})
		return
	}

	out := models.GroupsResponse{OK: resp.GetOk()}
	for _, item := range resp.GetData() {
		if mapped := toGroupModel(item); mapped != nil {
			h.resolveGroupDisplayName(mapped, actorID)
			out.Data = append(out.Data, *mapped)
		}
	}
	if resp.GetError() != nil {
		out.Error = &models.SendMessageError{
			Code:    resp.GetError().GetCode(),
			Message: resp.GetError().GetMessage(),
		}
	}

	status := http.StatusOK
	if !resp.GetOk() && resp.GetError() != nil {
		status = statusFromServiceCode(resp.GetError().GetCode(), http.StatusUnprocessableEntity)
	}
	respondJSON(w, status, out)
}

func (h *Handler) AddGroupMember(w http.ResponseWriter, r *http.Request) {
	conversationID, ok := groupIDFromPath(r)
	if !ok {
		respondJSON(w, http.StatusBadRequest, models.GroupMemberResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: invalidId},
		})
		return
	}

	var req models.AddGroupMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.GroupMemberResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: "invalid JSON"},
		})
		return
	}

	actorID := h.actorIDFromToken(r)
	if actorID == "" {
		respondJSON(w, http.StatusBadRequest, models.GroupMemberResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: "actor_id (or user_id / X-User-ID) required"},
		})
		return
	}
	if req.UserID == "" {
		respondJSON(w, http.StatusBadRequest, models.GroupMemberResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: "user_id required"},
		})
		return
	}

	protoReq := &apiv1.GroupAddMemberRequest{
		ActorId:        actorID,
		ConversationId: int32(conversationID),
		GroupId:        int32(conversationID),
		UserId:         req.UserID,
		Role:           int32(req.Role),
	}
	data, err := proto.Marshal(protoReq)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.GroupMemberResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "INTERNAL", Message: err.Error()},
		})
		return
	}

	reply, err := h.nc.Request(subjectGroupAddMember, data, requestTimeout)
	if err != nil {
		respondJSON(w, http.StatusBadGateway, models.GroupMemberResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "GATEWAY_ERROR", Message: "message-service unreachable: " + err.Error()},
		})
		return
	}

	var resp apiv1.GroupAddMemberResponse
	if err := proto.Unmarshal(reply.Data, &resp); err != nil {
		respondJSON(w, http.StatusBadGateway, models.GroupMemberResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "GATEWAY_ERROR", Message: "invalid response from message-service"},
		})
		return
	}

	out := models.GroupMemberResponse{OK: resp.GetOk()}
	if resp.GetData() != nil {
		out.Data = toGroupMemberModel(resp.GetData())
	}
	if resp.GetError() != nil {
		out.Error = &models.SendMessageError{
			Code:    resp.GetError().GetCode(),
			Message: resp.GetError().GetMessage(),
		}
	}

	status := http.StatusOK
	if !resp.GetOk() && resp.GetError() != nil {
		status = statusFromServiceCode(resp.GetError().GetCode(), http.StatusUnprocessableEntity)
	}
	respondJSON(w, status, out)
}

func (h *Handler) ListGroupMembers(w http.ResponseWriter, r *http.Request) {
	conversationID, ok := groupIDFromPath(r)
	if !ok {
		respondJSON(w, http.StatusBadRequest, models.GroupMembersResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: invalidId},
		})
		return
	}

	actorID := h.actorIDFromToken(r)
	if actorID == "" {
		respondJSON(w, http.StatusBadRequest, models.GroupMembersResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: "actor_id (or user_id / X-User-ID) required"},
		})
		return
	}

	protoReq := &apiv1.GroupListMembersRequest{
		ActorId:        actorID,
		ConversationId: int32(conversationID),
		GroupId:        int32(conversationID),
	}
	data, err := proto.Marshal(protoReq)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.GroupMembersResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "INTERNAL", Message: err.Error()},
		})
		return
	}

	reply, err := h.nc.Request(subjectGroupListMembers, data, requestTimeout)
	if err != nil {
		respondJSON(w, http.StatusBadGateway, models.GroupMembersResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "GATEWAY_ERROR", Message: "message-service unreachable: " + err.Error()},
		})
		return
	}

	var resp apiv1.GroupListMembersResponse
	if err := proto.Unmarshal(reply.Data, &resp); err != nil {
		respondJSON(w, http.StatusBadGateway, models.GroupMembersResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "GATEWAY_ERROR", Message: "invalid response from message-service"},
		})
		return
	}

	out := models.GroupMembersResponse{OK: resp.GetOk()}
	for _, item := range resp.GetData() {
		if mapped := toGroupMemberModel(item); mapped != nil {
			out.Data = append(out.Data, *mapped)
		}
	}
	if resp.GetError() != nil {
		out.Error = &models.SendMessageError{
			Code:    resp.GetError().GetCode(),
			Message: resp.GetError().GetMessage(),
		}
	}

	status := http.StatusOK
	if !resp.GetOk() && resp.GetError() != nil {
		status = statusFromServiceCode(resp.GetError().GetCode(), http.StatusUnprocessableEntity)
	}
	respondJSON(w, status, out)
}

func (h *Handler) UpdateGroupMemberRole(w http.ResponseWriter, r *http.Request) {
	conversationID, ok := groupIDFromPath(r)
	if !ok {
		respondJSON(w, http.StatusBadRequest, models.GroupMemberResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: invalidId},
		})
		return
	}
	targetUserID := chi.URLParam(r, "user_id")
	if targetUserID == "" {
		respondJSON(w, http.StatusBadRequest, models.GroupMemberResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: "user_id required"},
		})
		return
	}

	var req models.UpdateGroupMemberRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.GroupMemberResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: "invalid JSON"},
		})
		return
	}

	actorID := h.actorIDFromToken(r)
	if actorID == "" {
		respondJSON(w, http.StatusBadRequest, models.GroupMemberResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: "actor_id (or user_id / X-User-ID) required"},
		})
		return
	}

	protoReq := &apiv1.GroupUpdateRoleRequest{
		ActorId:        actorID,
		ConversationId: int32(conversationID),
		GroupId:        int32(conversationID),
		UserId:         targetUserID,
		Role:           int32(req.Role),
	}
	data, err := proto.Marshal(protoReq)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.GroupMemberResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "INTERNAL", Message: err.Error()},
		})
		return
	}

	reply, err := h.nc.Request(subjectGroupUpdateRole, data, requestTimeout)
	if err != nil {
		respondJSON(w, http.StatusBadGateway, models.GroupMemberResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "GATEWAY_ERROR", Message: "message-service unreachable: " + err.Error()},
		})
		return
	}

	var resp apiv1.GroupUpdateRoleResponse
	if err := proto.Unmarshal(reply.Data, &resp); err != nil {
		respondJSON(w, http.StatusBadGateway, models.GroupMemberResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "GATEWAY_ERROR", Message: "invalid response from message-service"},
		})
		return
	}

	out := models.GroupMemberResponse{OK: resp.GetOk()}
	if resp.GetData() != nil {
		out.Data = toGroupMemberModel(resp.GetData())
	}
	if resp.GetError() != nil {
		out.Error = &models.SendMessageError{
			Code:    resp.GetError().GetCode(),
			Message: resp.GetError().GetMessage(),
		}
	}

	status := http.StatusOK
	if !resp.GetOk() && resp.GetError() != nil {
		status = statusFromServiceCode(resp.GetError().GetCode(), http.StatusUnprocessableEntity)
	}
	respondJSON(w, status, out)
}

func (h *Handler) RemoveGroupMember(w http.ResponseWriter, r *http.Request) {
	conversationID, ok := groupIDFromPath(r)
	if !ok {
		respondJSON(w, http.StatusBadRequest, models.GroupResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: invalidId},
		})
		return
	}
	targetUserID := chi.URLParam(r, "user_id")
	if targetUserID == "" {
		respondJSON(w, http.StatusBadRequest, models.GroupResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: "user_id required"},
		})
		return
	}

	actorID := h.actorIDFromToken(r)
	if actorID == "" {
		respondJSON(w, http.StatusBadRequest, models.GroupResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: "actor_id (or user_id / X-User-ID) required"},
		})
		return
	}

	protoReq := &apiv1.GroupRemoveMemberRequest{
		ActorId:        actorID,
		ConversationId: int32(conversationID),
		GroupId:        int32(conversationID),
		UserId:         targetUserID,
	}
	data, err := proto.Marshal(protoReq)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.GroupResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "INTERNAL", Message: err.Error()},
		})
		return
	}

	reply, err := h.nc.Request(subjectGroupRemove, data, requestTimeout)
	if err != nil {
		respondJSON(w, http.StatusBadGateway, models.GroupResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "GATEWAY_ERROR", Message: "message-service unreachable: " + err.Error()},
		})
		return
	}

	var resp apiv1.GroupRemoveMemberResponse
	if err := proto.Unmarshal(reply.Data, &resp); err != nil {
		respondJSON(w, http.StatusBadGateway, models.GroupResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "GATEWAY_ERROR", Message: "invalid response from message-service"},
		})
		return
	}

	out := models.GroupResponse{OK: resp.GetOk()}
	if resp.GetError() != nil {
		out.Error = &models.SendMessageError{
			Code:    resp.GetError().GetCode(),
			Message: resp.GetError().GetMessage(),
		}
	}

	status := http.StatusOK
	if !resp.GetOk() && resp.GetError() != nil {
		status = statusFromServiceCode(resp.GetError().GetCode(), http.StatusUnprocessableEntity)
	}
	respondJSON(w, status, out)
}

func (h *Handler) LeaveGroup(w http.ResponseWriter, r *http.Request) {
	conversationID, ok := groupIDFromPath(r)
	if !ok {
		respondJSON(w, http.StatusBadRequest, models.GroupResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: invalidId},
		})
		return
	}

	actorID := h.actorIDFromToken(r)
	if actorID == "" {
		respondJSON(w, http.StatusBadRequest, models.GroupResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: "actor_id (or user_id / X-User-ID) required"},
		})
		return
	}

	protoReq := &apiv1.GroupLeaveRequest{
		UserId:         actorID,
		ConversationId: int32(conversationID),
		GroupId:        int32(conversationID),
	}
	data, err := proto.Marshal(protoReq)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.GroupResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "INTERNAL", Message: err.Error()},
		})
		return
	}

	reply, err := h.nc.Request(subjectGroupLeave, data, requestTimeout)
	if err != nil {
		respondJSON(w, http.StatusBadGateway, models.GroupResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "GATEWAY_ERROR", Message: "message-service unreachable: " + err.Error()},
		})
		return
	}

	var resp apiv1.GroupLeaveResponse
	if err := proto.Unmarshal(reply.Data, &resp); err != nil {
		respondJSON(w, http.StatusBadGateway, models.GroupResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "GATEWAY_ERROR", Message: "invalid response from message-service"},
		})
		return
	}

	out := models.GroupResponse{OK: resp.GetOk()}
	if resp.GetError() != nil {
		out.Error = &models.SendMessageError{
			Code:    resp.GetError().GetCode(),
			Message: resp.GetError().GetMessage(),
		}
	}

	status := http.StatusOK
	if !resp.GetOk() && resp.GetError() != nil {
		status = statusFromServiceCode(resp.GetError().GetCode(), http.StatusUnprocessableEntity)
	}
	respondJSON(w, status, out)
}

func (h *Handler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	conversationID, ok := groupIDFromPath(r)
	if !ok {
		respondJSON(w, http.StatusBadRequest, models.GroupResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: invalidId},
		})
		return
	}

	actorID := h.actorIDFromToken(r)
	if actorID == "" {
		respondJSON(w, http.StatusBadRequest, models.GroupResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: "actor_id (or user_id / X-User-ID) required"},
		})
		return
	}

	protoReq := &apiv1.GroupDeleteRequest{
		ActorId:        actorID,
		ConversationId: int32(conversationID),
		GroupId:        int32(conversationID),
	}
	data, err := proto.Marshal(protoReq)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.GroupResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "INTERNAL", Message: err.Error()},
		})
		return
	}

	reply, err := h.nc.Request(subjectGroupDelete, data, requestTimeout)
	if err != nil {
		respondJSON(w, http.StatusBadGateway, models.GroupResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "GATEWAY_ERROR", Message: "message-service unreachable: " + err.Error()},
		})
		return
	}

	var resp apiv1.GroupDeleteResponse
	if err := proto.Unmarshal(reply.Data, &resp); err != nil {
		respondJSON(w, http.StatusBadGateway, models.GroupResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "GATEWAY_ERROR", Message: "invalid response from message-service"},
		})
		return
	}

	out := models.GroupResponse{OK: resp.GetOk()}
	if resp.GetError() != nil {
		out.Error = &models.SendMessageError{
			Code:    resp.GetError().GetCode(),
			Message: resp.GetError().GetMessage(),
		}
	}

	status := http.StatusOK
	if !resp.GetOk() && resp.GetError() != nil {
		status = statusFromServiceCode(resp.GetError().GetCode(), http.StatusUnprocessableEntity)
	}
	respondJSON(w, status, out)
}

func toGroupModel(group *apiv1.Group) *models.Group {
	if group == nil {
		return nil
	}
	return &models.Group{
		ID:        int(group.GetId()),
		Name:      group.GetName(),
		AvatarURL: group.GetAvatarUrl(),
		CreatedBy: group.GetCreatedBy(),
		CreatedAt: group.GetCreatedAt(),
		UpdatedAt: group.GetUpdatedAt(),
	}
}

func toGroupMemberModel(member *apiv1.GroupMember) *models.GroupMember {
	if member == nil {
		return nil
	}
	conversationID := int(member.GetConversationId())
	return &models.GroupMember{
		ID:             int(member.GetId()),
		ConversationID: conversationID,
		GroupID:        conversationID,
		UserID:         member.GetUserId(),
		Role:           int(member.GetRole()),
		CreatedAt:      member.GetCreatedAt(),
	}
}

func groupIDFromPath(r *http.Request) (int, bool) {
	raw := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || id <= 0 {
		return 0, false
	}
	return int(id), true
}

// resolveGroupDisplayName résout le nom d'affichage d'une conversation si le nom en DB est vide.
// - Conv à 2 membres : nom de l'autre membre
// - Conv à 3+ membres : "user1, user2, user3"
// Si le nom est déjà renseigné (renommé par l'utilisateur), on le garde tel quel.
func (h *Handler) resolveGroupDisplayName(group *models.Group, actorID string) {
	if group == nil || group.Name != "" {
		return
	}

	members := h.fetchGroupMembers(group.ID, actorID)
	if len(members) == 0 {
		return
	}

	var names []string
	for _, m := range members {
		if m.UserID == actorID {
			continue
		}
		name := h.fetchUsername(m.UserID)
		if name != "" {
			names = append(names, name)
		}
	}

	if len(names) > 0 {
		group.Name = strings.Join(names, ", ")
	}
}

// fetchGroupMembers récupère les membres d'une conversation via NATS.
func (h *Handler) fetchGroupMembers(conversationID int, actorID string) []models.GroupMember {
	protoReq := &apiv1.GroupListMembersRequest{
		ActorId:        actorID,
		ConversationId: int32(conversationID),
		GroupId:        int32(conversationID),
	}
	data, err := proto.Marshal(protoReq)
	if err != nil {
		return nil
	}

	reply, err := h.nc.Request(subjectGroupListMembers, data, requestTimeout)
	if err != nil {
		return nil
	}

	var resp apiv1.GroupListMembersResponse
	if err := proto.Unmarshal(reply.Data, &resp); err != nil || !resp.GetOk() {
		return nil
	}

	members := make([]models.GroupMember, 0, len(resp.GetData()))
	for _, item := range resp.GetData() {
		if mapped := toGroupMemberModel(item); mapped != nil {
			members = append(members, *mapped)
		}
	}
	return members
}

// fetchUsername récupère le username d'un utilisateur via NATS (user service NestJS).
func (h *Handler) fetchUsername(userID string) string {
	request := struct {
		Pattern string            `json:"pattern"`
		Data    map[string]string `json:"data"`
		ID      string            `json:"id"`
	}{
		Pattern: "user.get",
		Data:    map[string]string{"id": userID},
		ID:      time.Now().String(),
	}
	payload, err := json.Marshal(request)
	if err != nil {
		return ""
	}

	msg, err := h.nc.Request("user.get", payload, 2*time.Second)
	if err != nil {
		log.Printf("[Gateway] fetchUsername: user.get error for %s: %v", userID, err)
		return ""
	}

	var wrapper struct {
		Response struct {
			Username    string `json:"username"`
			DisplayName string `json:"display_name"`
		} `json:"response"`
	}
	if err := json.Unmarshal(msg.Data, &wrapper); err != nil {
		return ""
	}

	if wrapper.Response.DisplayName != "" {
		return wrapper.Response.DisplayName
	}
	return wrapper.Response.Username
}
