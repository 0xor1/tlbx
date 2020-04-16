<template>
    <router-view/>
</template>

<style lang="scss">
@font-face {
    font-family: JB-Mono;
    src: url('~@/assets/font/JetBrainsMono-Regular.woff2') format('woff2');
    font-weight: normal;
    font-style: normal;
}
* {
    font-family: JB-Mono;
    font-weight: lighter;
    background-color: #000;
    color: #ddd;
    button {
        cursor: pointer;
        &:hover {
            background-color: #555;
        }
        &:active {
            background-color: #222;
        }
    }
    input, button {
        background-color: #222;
        outline: none;
        border: 1px solid #fff;
        border-radius: 2px;
        padding: 5px;
        &::placeholder{
            color: #aaa;
        }
    }
}
</style>