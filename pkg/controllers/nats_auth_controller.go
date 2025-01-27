package controllers

import (
	"context"
	"errors"
	"fmt"
	"github.com/mynaparrot/plugnmeet-server/pkg/config"
	"github.com/mynaparrot/plugnmeet-server/pkg/models"
	"github.com/mynaparrot/plugnmeet-server/pkg/services/nats"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats.go/micro"
	"github.com/nats-io/nkeys"
	log "github.com/sirupsen/logrus"
)

type NatsAuthController struct {
	ctx           context.Context
	app           *config.AppConfig
	authModel     *models.AuthModel
	natsService   *natsservice.NatsService
	issuerKeyPair nkeys.KeyPair
}

func NewNatsAuthController(app *config.AppConfig, authModel *models.AuthModel, kp nkeys.KeyPair) *NatsAuthController {
	return &NatsAuthController{
		ctx:           context.Background(),
		app:           app,
		authModel:     authModel,
		natsService:   natsservice.New(app),
		issuerKeyPair: kp,
	}
}

func (s *NatsAuthController) Handle(r micro.Request) {
	rc, err := jwt.DecodeAuthorizationRequestClaims(string(r.Data()))
	if err != nil {
		log.Println("Error", err)
		_ = r.Error("500", err.Error(), nil)
	}
	userNkey := rc.UserNkey
	serverId := rc.Server.ID

	claims, err := s.handleClaims(rc)
	if err != nil {
		s.Respond(r, userNkey, serverId, "", err)
		return
	}

	token, err := ValidateAndSign(claims, s.issuerKeyPair)
	s.Respond(r, userNkey, serverId, token, err)
}

func (s *NatsAuthController) handleClaims(req *jwt.AuthorizationRequestClaims) (*jwt.UserClaims, error) {
	claims := jwt.NewUserClaims(req.UserNkey)
	claims.Audience = s.app.NatsInfo.Account

	// check the info first
	data, err := s.authModel.VerifyPlugNmeetAccessToken(req.ConnectOptions.Token)
	if err != nil {
		return nil, err
	}

	roomId := data.GetRoomId()
	userId := data.GetUserId()

	userInfo, err := s.natsService.GetUserInfo(roomId, userId)
	if err != nil {
		return nil, err
	}
	if userInfo == nil {
		return nil, errors.New(fmt.Sprintf("User info not found for userId: %s, roomId: %s", userId, roomId))
	}

	allow := jwt.StringList{
		"$JS.API.INFO",
		fmt.Sprintf("$JS.API.STREAM.INFO.%s", roomId),
		// allow sending messages to the system
		fmt.Sprintf("%s.%s.%s", s.app.NatsInfo.Subjects.SystemJsWorker, roomId, userId),
	}

	chatPermission, err := s.natsService.CreateChatConsumer(roomId, userId)
	if err != nil {
		return nil, err
	}
	allow.Add(chatPermission...)

	sysPublicPermission, err := s.natsService.CreateSystemPublicConsumer(roomId, userId)
	if err != nil {
		return nil, err
	}
	allow.Add(sysPublicPermission...)

	sysPrivatePermission, err := s.natsService.CreateSystemPrivateConsumer(roomId, userId)
	if err != nil {
		return nil, err
	}
	allow.Add(sysPrivatePermission...)

	whiteboardPermission, err := s.natsService.CreateWhiteboardConsumer(roomId, userId)
	if err != nil {
		return nil, err
	}
	allow.Add(whiteboardPermission...)

	dataChannelPermission, err := s.natsService.CreateDataChannelConsumer(roomId, userId)
	if err != nil {
		return nil, err
	}
	allow.Add(dataChannelPermission...)

	// put name in proper format
	claims.Name = fmt.Sprintf("%s:%s", roomId, userId)
	// Assign Permissions
	claims.Permissions = jwt.Permissions{
		Pub: jwt.Permission{
			Allow: allow,
		},
	}

	return claims, nil
}

func (s *NatsAuthController) Respond(req micro.Request, userNKey, serverId, userJWT string, err error) {
	rc := jwt.NewAuthorizationResponseClaims(userNKey)
	rc.Audience = serverId
	rc.Jwt = userJWT
	if err != nil {
		rc.Error = err.Error()
	}

	token, err := rc.Encode(s.issuerKeyPair)
	if err != nil {
		log.Errorln("error encoding response jwt:", err)
	}

	_ = req.Respond([]byte(token))
}

func ValidateAndSign(claims *jwt.UserClaims, kp nkeys.KeyPair) (string, error) {
	// Validate the claims.
	vr := jwt.CreateValidationResults()
	claims.Validate(vr)
	if len(vr.Errors()) > 0 {
		return "", errors.Join(vr.Errors()...)
	}

	// Sign it with the issuer key since this is non-operator mode.
	return claims.Encode(kp)
}
