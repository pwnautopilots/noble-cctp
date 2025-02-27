// Copyright 2024 Circle Internet Group, Inc.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/circlefin/noble-cctp/x/cctp/types"
	"github.com/ethereum/go-ethereum/common"
)

func (k msgServer) DisableAttester(goCtx context.Context, msg *types.MsgDisableAttester) (*types.MsgDisableAttesterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	attesterManager := k.GetAttesterManager(ctx)
	if attesterManager != msg.From {
		return nil, errors.Wrapf(types.ErrUnauthorized, "this message sender cannot disable attesters")
	}

	if len(common.FromHex(msg.Attester)) == 0 {
		return nil, errors.Wrapf(types.ErrInvalidAddress, "invalid attester")
	}

	_, found := k.GetAttester(ctx, msg.Attester)
	if !found {
		return nil, errors.Wrapf(types.ErrDisableAttester, "attester not found")
	}

	// disallow removing public key if there is only 1 active public key
	storedAttesters := k.GetAllAttesters(ctx)
	if len(storedAttesters) == 1 {
		return nil, errors.Wrap(types.ErrDisableAttester, "cannot disable the last attester")
	}

	// disallow removing public key if it causes the n in m/n multisig to fall below m (threshold # of signers)
	signatureThreshold, found := k.GetSignatureThreshold(ctx)
	if !found {
		return nil, errors.Wrapf(types.ErrDisableAttester, "signature threshold not set")
	}

	if uint32(len(storedAttesters)) <= signatureThreshold.Amount {
		return nil, errors.Wrap(types.ErrDisableAttester, "signature threshold is too low")
	}

	k.DeleteAttester(ctx, msg.Attester)

	event := types.AttesterDisabled{
		Attester: msg.Attester,
	}
	err := ctx.EventManager().EmitTypedEvent(&event)

	return &types.MsgDisableAttesterResponse{}, err
}
