package service

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	kitlog "github.com/go-kit/kit/log"
	"github.com/igm/sockjs-go/v3/sockjs"
	"github.com/kolide/fleet/server/contexts/viewer"
	"github.com/kolide/fleet/server/kolide"
	"github.com/kolide/fleet/server/websocket"
)

////////////////////////////////////////////////////////////////////////////////
// Create Distributed Query Campaign
////////////////////////////////////////////////////////////////////////////////

type createDistributedQueryCampaignRequest struct {
	Query    string                          `json:"query"`
	Selected distributedQueryCampaignTargets `json:"selected"`
}

type distributedQueryCampaignTargets struct {
	Labels []uint `json:"labels"`
	Hosts  []uint `json:"hosts"`
}

type createDistributedQueryCampaignResponse struct {
	Campaign *kolide.DistributedQueryCampaign `json:"campaign,omitempty"`
	Err      error                            `json:"error,omitempty"`
}

func (r createDistributedQueryCampaignResponse) error() error { return r.Err }

func makeCreateDistributedQueryCampaignEndpoint(svc kolide.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createDistributedQueryCampaignRequest)
		campaign, err := svc.NewDistributedQueryCampaign(ctx, req.Query, req.Selected.Hosts, req.Selected.Labels)
		if err != nil {
			return createDistributedQueryCampaignResponse{Err: err}, nil
		}
		return createDistributedQueryCampaignResponse{Campaign: campaign}, nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// Create Distributed Query Campaign By Names
////////////////////////////////////////////////////////////////////////////////

type createDistributedQueryCampaignByNamesRequest struct {
	Query    string                                 `json:"query"`
	Selected distributedQueryCampaignTargetsByNames `json:"selected"`
}

type distributedQueryCampaignTargetsByNames struct {
	Labels []string `json:"labels"`
	Hosts  []string `json:"hosts"`
}

func makeCreateDistributedQueryCampaignByNamesEndpoint(svc kolide.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createDistributedQueryCampaignByNamesRequest)
		campaign, err := svc.NewDistributedQueryCampaignByNames(ctx, req.Query, req.Selected.Hosts, req.Selected.Labels)
		if err != nil {
			return createDistributedQueryCampaignResponse{Err: err}, nil
		}
		return createDistributedQueryCampaignResponse{Campaign: campaign}, nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// Stream Distributed Query Campaign Results and Metadata
////////////////////////////////////////////////////////////////////////////////

func makeStreamDistributedQueryCampaignResultsHandler(svc kolide.Service, jwtKey string, logger kitlog.Logger) http.Handler {
	opt := sockjs.DefaultOptions
	opt.Websocket = true
	opt.RawWebsocket = true
	opt.JSessionID = func(rw http.ResponseWriter, req *http.Request) {
		// Set the JSESSIONID cookie that will help some load balancers to
		// maintain sticky sessions (needed for SockJS in XHR mode)
		cookie, err := req.Cookie("JSESSIONID")
		if err == http.ErrNoCookie {
			id, err := kolide.RandomText(8)
			if err != nil {
				logger.Log("err", err, "msg", "generate JSESSIONID")
				id = ""
			}
			cookie = &http.Cookie{
				Name:  "JSESSIONID",
				Value: id,
			}
		}
		cookie.Path = "/"
		// Setting a 30 second Max-Age means that the session will be sticky as
		// long as the query keeps running (because this cookie will be
		// refreshed on each request), but will allow a new session to be
		// created after some time passes between requests.
		cookie.MaxAge = 30
		header := rw.Header()
		header.Add("Set-Cookie", cookie.String())
	}
	return sockjs.NewHandler("/api/v1/kolide/results", opt, func(session sockjs.Session) {
		defer session.Close(0, "none")

		conn := &websocket.Conn{Session: session}

		// Receive the auth bearer token
		token, err := conn.ReadAuthToken()
		if err != nil {
			logger.Log("err", err, "msg", "failed to read auth token")
			return
		}

		// Authenticate with the token
		vc, err := authViewer(context.Background(), jwtKey, token, svc)
		if err != nil || !vc.CanPerformActions() {
			logger.Log("err", err, "msg", "unauthorized viewer")
			conn.WriteJSONError("unauthorized")
			return
		}

		ctx := viewer.NewContext(context.Background(), *vc)

		msg, err := conn.ReadJSONMessage()
		if err != nil {
			logger.Log("err", err, "msg", "reading select_campaign JSON")
			conn.WriteJSONError("error reading select_campaign")
			return
		}
		if msg.Type != "select_campaign" {
			logger.Log("err", "unexpected msg type, expected select_campaign", "msg-type", msg.Type)
			conn.WriteJSONError("expected select_campaign")
			return
		}

		var info struct {
			CampaignID uint `json:"campaign_id"`
		}
		err = json.Unmarshal(*(msg.Data.(*json.RawMessage)), &info)
		if err != nil {
			logger.Log("err", err, "msg", "unmarshaling select_campaign data")
			conn.WriteJSONError("error unmarshaling select_campaign data")
			return
		}
		if info.CampaignID == 0 {
			logger.Log("err", "campaign ID not set")
			conn.WriteJSONError("0 is not a valid campaign ID")
			return
		}

		svc.StreamCampaignResults(ctx, conn, info.CampaignID)

	})
}
