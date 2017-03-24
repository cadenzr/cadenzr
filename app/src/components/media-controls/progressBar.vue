<template>
    <div ref="progressBarContainer" class="progress-bar-container">
        <div class="progress-bar-padding"></div>
        <div ref="progressBar" class="progress-bar"></div>

        <div ref="progressBarHover" class="progress-bar-hover"></div>
        <div ref="progressBarPlayed" class="progress-bar-played"></div>
        <div ref="progressBarScrubber" class="progress-bar-scrubber"></div>
    </div>
</template>

<script lang="ts">
    import Vue from 'vue';
    import * as _ from 'lodash';
    import PubSub from '@/PubSub';
    import {events as AudioPlayerEvents} from '@/AudioPlayer';
    import AudioPlayer from '@/AudioPlayer';

    interface ProgressBar extends Vue {
        width: number;
        subscriptions: Array<any>;
    }

    export default {
        name: 'progress-bar',
            data: function () {
                return {
                    width: 0,
                    subscriptions: [],
                }
            },
            mounted: function () {
                let self = this;


                let scrubberEl = (<any>this).$refs.progressBarScrubber;
                let progressBarHoverEl = (<any>this).$refs.progressBarHover;
                let progressBarContainerEl = (<any>this).$refs.progressBarContainer;
                let progressBarPlayedEl = (<any>this).$refs.progressBarPlayed;

                let scrubberComputedStyle = <any>window.getComputedStyle(scrubberEl);
                let scrubberSize = Number(scrubberComputedStyle.width.replace('px' ,''));

                let rect = (<any>this).$refs.progressBar.getBoundingClientRect();
                this.width = rect.width;
                let scrubberX = 0;

                let isSeeking = false;

                let isPlaying = false;
                let timeLeft = 0;

                (<any>self).subscriptions.push(PubSub.subscribe(AudioPlayerEvents.Play, (song) => {
                    isPlaying = true;
                    let played = (<any>AudioPlayer).currentSongTime();
                    timeLeft = song.duration - played;
                    let progress = played / song.duration;

                    setScrubberPosition(progress * rect.width);
                    prevTimestamp = 0;
                    requestAnimationFrame(renderScrubber);
                }));

                (<any>self).subscriptions.push(PubSub.subscribe(AudioPlayerEvents.Pause, () => {
                    isPlaying = false;
                }));



                let prevTimestamp = 0;
                let renderScrubber = (timestamp) => {
                    if(isPlaying && isSeeking) {
                        return;
                    }

                    let dt = 0;
                    if(prevTimestamp === 0) {
                        dt = 0;
                        prevTimestamp = timestamp;
                    } else {
                        dt = (timestamp - prevTimestamp)/1000;
                        prevTimestamp = timestamp;
                    }

                    let increase = ((rect.width-scrubberX)/timeLeft) * dt;
                    setScrubberPosition(scrubberX + increase);
                    setProgressPlayedWidth(scrubberX + increase);

                    if(isPlaying) {
                        window.requestAnimationFrame(renderScrubber);
                    }
                };

                let setScrubberPosition = (x) => {
                    scrubberX = x;
                    let shifted = (scrubberX - scrubberSize/2);
                    scrubberEl.style.transform = 'translate(' + shifted.toString() + 'px)';

                };

                let setProgressHoverWidth = (x) => {
                    progressBarHoverEl.style.width = x.toString() + 'px';
                };

                let setProgressPlayedWidth = (x) => {
                    progressBarPlayedEl.style.width = x.toString() + 'px';
                };

                let mousePositionToScrubber = (x) => {
                    let relativeX = x - rect.left;
                    relativeX = Math.max(0, Math.min(relativeX, rect.width));
                    return relativeX;
                };

                let scrubberMove = (e) => {
                    let absoluteX = getMousePosition(e);
                    //alert(e.touches[0].clientX);
                    let scrubberX = mousePositionToScrubber(absoluteX);
                    setScrubberPosition(scrubberX);
                    setProgressPlayedWidth(scrubberX);
                };

                let scrubberReleased = (e) => {
                    isSeeking = false;
                    window.removeEventListener('mousemove', scrubberMove);
                    window.removeEventListener('mouseup', scrubberReleased);
                    window.removeEventListener('touchmove', scrubberMove);
                    window.removeEventListener('touchend', scrubberReleased);

                    let progress = scrubberX / rect.width;
                    AudioPlayer.seekFromPercentage(progress);
                    AudioPlayer.play();
                };

                let scrubberStartMove = (e) => {
                    e.preventDefault();
                    isSeeking = true;
                    window.addEventListener('mousemove', scrubberMove);
                    window.addEventListener('mouseup', scrubberReleased);
                    window.addEventListener('touchmove', scrubberMove);
                    window.addEventListener('touchend', scrubberReleased);
                };

                let getMousePosition = (e) => {
                    if(e.clientX !== undefined) {
                        return e.clientX;
                    } else if(e.touches) {
                        return e.touches[0].clientX;
                    }

                    return 0;
                }

                (<any>this).$refs.progressBarScrubber.addEventListener('mousedown', (e) => {
                    scrubberStartMove(e);
                });



                (<any>this).$refs.progressBarContainer.addEventListener('mousedown', (e) => {                    
                    let scrubberX = mousePositionToScrubber(getMousePosition(e));

                    scrubberEl.style.transition = '';
                    progressBarPlayedEl.style.transition = '';

                    setScrubberPosition(scrubberX);
                    setProgressPlayedWidth(scrubberX);
                    scrubberStartMove(e);
                });

                (<any>this).$refs.progressBarContainer.addEventListener('touchstart', (e) => {
                    let scrubberX = mousePositionToScrubber(getMousePosition(e));
                    setScrubberPosition(scrubberX);
                    setProgressPlayedWidth(scrubberX);
                    scrubberStartMove(e);
                });

                (<any>this).$refs.progressBarContainer.addEventListener('mousemove', (e) => {
                    let scrubberX = mousePositionToScrubber(getMousePosition(e));
                    setProgressHoverWidth(scrubberX);
                });

                (<any>this).$refs.progressBarContainer.addEventListener('mouseout', (e) => {
                    setProgressHoverWidth(0);
                });
            },
            beforeDestroy () {
                _.forEach((<any>this).subscriptions, (s:any) => {
                    PubSub.unsubscribe(s);
                });
            },
            methods: {

            }
    } as Vue.ComponentOptions<ProgressBar>;
</script>


<style lang="scss">
$progress-bar-height: 3px;
$progress-bar-active-height: 6px;
$progress-bar-color: #8e8e8e;
$progress-bar-hover-color: rgba(255,255,255,.5);
$progress-bar-played-color: red;
$progress-bar-scrubber-size: 15px;

.progress-bar-container {
    width: 100%;
    position: relative;
    height: $progress-bar-height;
    cursor: pointer;
    outline:none;

    .progress-bar-padding {
        $padding: 16px;
        width: 100%;
        height: 100%;
        padding-top: 2*$padding;
        position: absolute;
        bottom: -$padding;
        z-index: 1001;
    }

    .progress-bar-scrubber {
        opacity: 0;
        width: $progress-bar-scrubber-size;
        height: $progress-bar-scrubber-size;
        border-radius: 50%;
        background-color: red;
        position: absolute;
        transform: translate(-$progress-bar-scrubber-size/2);
        bottom: 0px + $progress-bar-active-height/2 - $progress-bar-scrubber-size/2;
        z-index: 1000;   
    }

    .progress-bar {
        top: 0px;
        height: 100%;
        //transition: height 0.1s;
        position: absolute;
        bottom: 0px;
        background-color: $progress-bar-color;
        width: 100%;
        z-index: 997;
        transition: transform 0.2s;
    }

    .progress-bar-hover {
        height: 100%;
        position: absolute;
        width: 0px;
        background-color: $progress-bar-hover-color;
        z-index: 998;
        transition: transform 0.2s;

    }

    .progress-bar-played {
        height: 100%;
        position: absolute;
        width: 0px;
        background-color: $progress-bar-played-color;
        z-index: 999;
        transition: transform 0.2s;

    }
}


.progress-bar-container:hover {


    .progress-bar, .progress-bar-hover, .progress-bar-played {
        transform: scaleY(2);
    }

    .progress-bar-scrubber {
        display: block;
        opacity: 1;
    }
    

}



</style>