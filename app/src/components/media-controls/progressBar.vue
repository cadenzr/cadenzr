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

    interface ProgressBar extends Vue {
        width: number;
    }

    export default {
        name: 'progress-bar',
            data: function () {
                return {
                    width: 0,
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

                this.$on('progress-bar-start', (s) => {
                    scrubberEl.style.transition = 'transform ' + s.duration + 's';
                    scrubberEl.style.transform = 'translate(' + rect.width + 'px)';

                    progressBarPlayedEl.style.transition = 'width ' + s.duration + 's';
                    progressBarPlayedEl.style.width = rect.width + 'px';
                });

                let setScrubberPosition = (x) => {
                    scrubberEl.style.transform = 'translate(' + (x - scrubberSize/2).toString() + 'px)';
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
                    window.removeEventListener('mousemove', scrubberMove);
                    window.removeEventListener('touchmove', scrubberMove);
                    // TODO remove ourself?
                };

                let scrubberStartMove = (e) => {
                    e.preventDefault();
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