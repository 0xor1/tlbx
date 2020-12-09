<template>
    <div>
        <loginout></loginout>
        <router-view/>
    </div>
</template>

<script>
import loginout from './components/loginout'
  export default {
    name: 'root',
    components: { loginout },
  }
</script>

<style lang="scss">
@import url(https://fonts.googleapis.com/css2?family=Roboto+Mono:wght@400;700&display=swap);
* {
    font-family: 'Roboto Mono', monospace;
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