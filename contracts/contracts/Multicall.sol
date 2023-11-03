// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.9;

contract Multicall {
    function getPricesTypeOne(
        address[] calldata pools
    ) external view returns (int24[] memory) {
        int24[] memory ticks = new int24[](pools.length);
        for (uint i = 0; i < pools.length; i++) {
            (, bytes memory data) = pools[i].staticcall(
                abi.encodeWithSignature("slot0()")
            );
            (, ticks[i], , , , , ) = abi.decode(
                data,
                (uint160, int24, uint16, uint16, uint16, uint8, bool)
            );
        }
        return ticks;
    }

    function getPriceTypeTwo(
        address[] calldata pools
    ) external view returns (uint112[] memory, uint112[] memory) {
        uint112[] memory x = new uint112[](pools.length);
        uint112[] memory y = new uint112[](pools.length);

        for (uint i = 0; i < pools.length; i++) {
            (, bytes memory data) = pools[i].staticcall(
                abi.encodeWithSignature("getReserves()")
            );
            (x[i], y[i], ) = abi.decode(data, (uint112, uint112, uint32));
        }
        return (x, y);
    }
}
