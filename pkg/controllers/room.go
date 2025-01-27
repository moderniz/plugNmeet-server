package controllers

import (
	"github.com/bufbuild/protovalidate-go"
	"github.com/gofiber/fiber/v2"
	"github.com/mynaparrot/plugnmeet-protocol/plugnmeet"
	"github.com/mynaparrot/plugnmeet-protocol/utils"
	"github.com/mynaparrot/plugnmeet-server/pkg/models"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func HandleRoomCreate(c *fiber.Ctx) error {
	op := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
	req := new(plugnmeet.CreateRoomReq)
	err := op.Unmarshal(c.Body(), req)
	if err != nil {
		return utils.SendCommonProtoJsonResponse(c, false, err.Error())
	}

	v, err := protovalidate.New()
	if err != nil {
		return utils.SendCommonProtoJsonResponse(c, false, "failed to initialize validator: "+err.Error())
	}

	if err = v.Validate(req); err != nil {
		return utils.SendCommonProtoJsonResponse(c, false, err.Error())
	}

	m := models.NewRoomModel(nil, nil, nil)
	room, err := m.CreateRoom(req)
	if err != nil {
		return utils.SendCommonProtoJsonResponse(c, false, err.Error())
	}

	r := &plugnmeet.CreateRoomRes{
		Status:   true,
		Msg:      "success",
		RoomInfo: room,
	}

	return utils.SendProtoJsonResponse(c, r)
}

func HandleIsRoomActive(c *fiber.Ctx) error {
	req := new(plugnmeet.IsRoomActiveReq)
	op := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
	err := op.Unmarshal(c.Body(), req)
	if err != nil {
		return utils.SendCommonProtoJsonResponse(c, false, err.Error())
	}

	v, err := protovalidate.New()
	if err != nil {
		return utils.SendCommonProtoJsonResponse(c, false, "failed to initialize validator: "+err.Error())
	}

	if err = v.Validate(req); err != nil {
		return utils.SendCommonProtoJsonResponse(c, false, err.Error())
	}

	m := models.NewRoomModel(nil, nil, nil)
	res, _ := m.IsRoomActive(req)
	return utils.SendProtoJsonResponse(c, res)
}

func HandleGetActiveRoomInfo(c *fiber.Ctx) error {
	req := new(plugnmeet.GetActiveRoomInfoReq)
	op := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
	err := op.Unmarshal(c.Body(), req)
	if err != nil {
		return c.JSON(fiber.Map{
			"status": false,
			"msg":    err.Error(),
		})
	}
	v, err := protovalidate.New()
	if err != nil {
		return utils.SendCommonProtoJsonResponse(c, false, "failed to initialize validator: "+err.Error())
	}

	if err = v.Validate(req); err != nil {
		return utils.SendCommonProtoJsonResponse(c, false, err.Error())
	}
	m := models.NewRoomModel(nil, nil, nil)
	status, msg, res := m.GetActiveRoomInfo(req)

	r := &plugnmeet.GetActiveRoomInfoRes{
		Status: status,
		Msg:    msg,
		Room:   res,
	}

	return utils.SendProtoJsonResponse(c, r)
}

func HandleGetActiveRoomsInfo(c *fiber.Ctx) error {
	m := models.NewRoomModel(nil, nil, nil)
	status, msg, res := m.GetActiveRoomsInfo()

	r := &plugnmeet.GetActiveRoomsInfoRes{
		Status: status,
		Msg:    msg,
		Rooms:  res,
	}

	return utils.SendProtoJsonResponse(c, r)
}

func HandleEndRoom(c *fiber.Ctx) error {
	req := new(plugnmeet.RoomEndReq)
	op := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
	err := op.Unmarshal(c.Body(), req)
	if err != nil {
		return utils.SendCommonProtoJsonResponse(c, false, err.Error())
	}
	v, err := protovalidate.New()
	if err != nil {
		return utils.SendCommonProtoJsonResponse(c, false, "failed to initialize validator: "+err.Error())
	}

	if err = v.Validate(req); err != nil {
		return utils.SendCommonProtoJsonResponse(c, false, err.Error())
	}

	m := models.NewRoomModel(nil, nil, nil)
	status, msg := m.EndRoom(req)

	return utils.SendCommonProtoJsonResponse(c, status, msg)
}

func HandleFetchPastRooms(c *fiber.Ctx) error {
	req := new(plugnmeet.FetchPastRoomsReq)
	op := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
	err := op.Unmarshal(c.Body(), req)
	if err != nil {
		return utils.SendCommonProtoJsonResponse(c, false, err.Error())
	}
	v, err := protovalidate.New()
	if err != nil {
		return utils.SendCommonProtoJsonResponse(c, false, "failed to initialize validator: "+err.Error())
	}

	if err = v.Validate(req); err != nil {
		return utils.SendCommonProtoJsonResponse(c, false, err.Error())
	}

	m := models.NewRoomModel(nil, nil, nil)
	result, err := m.FetchPastRooms(req)

	if err != nil {
		return utils.SendCommonProtoJsonResponse(c, false, err.Error())
	}
	if result.GetTotalRooms() == 0 {
		return utils.SendCommonProtoJsonResponse(c, false, "no info found")
	}

	r := &plugnmeet.FetchPastRoomsRes{
		Status: true,
		Msg:    "success",
		Result: result,
	}
	return utils.SendProtoJsonResponse(c, r)
}

func HandleEndRoomForAPI(c *fiber.Ctx) error {
	isAdmin := c.Locals("isAdmin")
	roomId := c.Locals("roomId")

	if !isAdmin.(bool) {
		return utils.SendCommonProtobufResponse(c, false, "only admin can perform this task")
	}

	req := new(plugnmeet.RoomEndReq)
	err := proto.Unmarshal(c.Body(), req)
	if err != nil {
		return utils.SendCommonProtobufResponse(c, false, err.Error())
	}

	if roomId != req.RoomId {
		return utils.SendCommonProtobufResponse(c, false, "requested roomId & token roomId mismatched")
	}

	m := models.NewRoomModel(nil, nil, nil)
	status, msg := m.EndRoom(req)
	return utils.SendCommonProtobufResponse(c, status, msg)
}

func HandleChangeVisibilityForAPI(c *fiber.Ctx) error {
	isAdmin := c.Locals("isAdmin")
	roomId := c.Locals("roomId")

	if !isAdmin.(bool) {
		return utils.SendCommonProtobufResponse(c, false, "only admin can perform this task")
	}

	req := new(plugnmeet.ChangeVisibilityRes)
	err := proto.Unmarshal(c.Body(), req)
	if err != nil {
		return utils.SendCommonProtobufResponse(c, false, err.Error())
	}

	if roomId != req.RoomId {
		return utils.SendCommonProtobufResponse(c, false, "requested roomId & token roomId mismatched")
	}

	m := models.NewRoomModel(nil, nil, nil)
	status, msg := m.ChangeVisibility(req)
	return utils.SendCommonProtobufResponse(c, status, msg)
}
