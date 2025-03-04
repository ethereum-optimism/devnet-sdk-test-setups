// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { ERC20 } from "solady/tokens/ERC20.sol";

contract MockERC20 is ERC20 {
    function name() public pure override returns (string memory) {
        return "MockERC20";
    }

    function symbol() public pure override returns (string memory) {
        return "MCK";
    }

    function mint(address to, uint256 amount) external {
        _mint(to, amount);
    }
}