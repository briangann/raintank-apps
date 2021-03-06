package api

import (
	"fmt"

	"github.com/raintank/raintank-apps/task-server/api/rbody"
	"github.com/raintank/raintank-apps/task-server/model"
	"github.com/raintank/raintank-apps/task-server/sqlstore"
	"github.com/raintank/worldping-api/pkg/log"
)

func GetAgents(ctx *Context, query model.GetAgentsQuery) {
	query.OrgId = ctx.OrgId
	agents, err := sqlstore.GetAgents(&query)
	if err != nil {
		log.Error(3, err.Error())
		ctx.JSON(200, rbody.ErrResp(500, err))
		return
	}
	ctx.JSON(200, rbody.OkResp("agents", agents))
}

func GetAgentById(ctx *Context) {
	id := ctx.ParamsInt64(":id")
	owner := ctx.OrgId
	agent, err := sqlstore.GetAgentById(id, owner)
	if err == model.AgentNotFound {
		ctx.JSON(200, rbody.ErrResp(404, fmt.Errorf("agent not found")))
		return
	}
	if err != nil {
		log.Error(3, err.Error())
		ctx.JSON(200, rbody.ErrResp(500, err))
		return
	}

	ctx.JSON(200, rbody.OkResp("agent", agent))
}

func GetAgentMetrics(ctx *Context) {
	id := ctx.ParamsInt64(":id")
	owner := ctx.OrgId
	agent, err := sqlstore.GetAgentById(id, owner)
	if err != nil {
		log.Error(3, err.Error())
		ctx.JSON(200, rbody.ErrResp(500, err))
		return
	}
	if agent == nil {
		ctx.JSON(200, rbody.ErrResp(404, fmt.Errorf("agent not found")))
		return
	}
	metrics, err := sqlstore.GetAgentMetrics(agent)
	if err != nil {
		log.Error(3, err.Error())
		ctx.JSON(200, rbody.ErrResp(500, err))
		return
	}

	ctx.JSON(200, rbody.OkResp("metrics", metrics))
}

func AddAgent(ctx *Context, agent model.AgentDTO) {
	if !agent.ValidName() {
		ctx.JSON(400, "invalde agent Name. must match /^[0-9a-Z_-]+$/")
		return
	}
	agent.Id = 0
	//need to add suport for middelware context with AUTH/
	agent.OrgId = ctx.OrgId
	err := sqlstore.AddAgent(&agent)
	if err != nil {
		log.Error(3, err.Error())
		ctx.JSON(200, rbody.ErrResp(500, err))
		return
	}
	ctx.JSON(200, rbody.OkResp("agent", agent))
}

func UpdateAgent(ctx *Context, agent model.AgentDTO) {
	if !agent.ValidName() {
		ctx.JSON(200, rbody.ErrResp(400, fmt.Errorf("invalid agent Name. must match /^[0-9a-Z_-]+$/")))
		return
	}
	if agent.Id == 0 {
		ctx.JSON(200, rbody.ErrResp(400, fmt.Errorf("agent ID not set.")))
		return
	}
	//need to add suport for middelware context with AUTH/
	agent.OrgId = ctx.OrgId
	err := sqlstore.UpdateAgent(&agent)
	if err != nil {
		log.Error(3, err.Error())
		ctx.JSON(200, rbody.ErrResp(500, err))
		return
	}
	ctx.JSON(200, rbody.OkResp("agent", agent))
}

func DeleteAgent(ctx *Context) {
	id := ctx.ParamsInt64(":id")
	owner := ctx.OrgId
	err := sqlstore.DeleteAgent(id, owner)
	if err != nil {
		if err == model.AgentNotFound {
			ctx.JSON(200, rbody.ErrResp(404, fmt.Errorf("agent not found")))
			return
		}
		log.Error(3, err.Error())
		ctx.JSON(200, rbody.ErrResp(500, err))
		return
	}

	ActiveSockets.CloseSocketByAgentId(id)

	ctx.JSON(200, rbody.OkResp("agent", nil))
}
