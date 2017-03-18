
<template>
        <audio-player ref="player"></audio-player>
        
        <div class="pure-g">
            <div class="pure-u-4-24">
                <sidebar></sidebar>
            </div>
            
            <div class="pure-u-16-24">
                <transition name="fade" mode="out-in">
                    <router-view></router-view>
                </transition>
            </div>
            
            <div class="pure-u-4-24">
            </div>
            
        </div>
        
        <div class="audio-player-placeholder">
            
        </div>
</template>

<script lang="ts">
import Vue from "vue";


export default {
  name: "app",
} as Vue.ComponentOptions<Vue>;
</script>