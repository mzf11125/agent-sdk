// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import {IAgentVerifiable} from "@agent-ercs/verify/ERC8274/IAgentVerifiable.sol";

/// @title MockAgentVerifiable
/// @notice Reference implementation of IAgentVerifiable for local testing
///         only — stands in for a settlement contract that declares which
///         IAgentVerifier it trusts.
contract MockAgentVerifiable is IAgentVerifiable {
    address private _agentVerifier;

    constructor(address initialVerifier) {
        _agentVerifier = initialVerifier;
    }

    function agentVerifier() external view override returns (address) {
        return _agentVerifier;
    }

    function setAgentVerifier(address newVerifier) external {
        emit AgentVerifierUpdated(_agentVerifier, newVerifier);
        _agentVerifier = newVerifier;
    }
}
