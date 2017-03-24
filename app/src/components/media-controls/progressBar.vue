<template>
    <div class="progress-bar-container">
        <div ref="progressBar" class="progress-bar">
            <div ref="progressBarHover" class="progress-bar-hover"></div>
            <div ref="progressBarPlayed" class="progress-bar-played"></div>
            <div ref="progressBarScrubber" class="progress-bar-scrubber"></div>
        </div>
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



                (<any>this).$refs.progressBar.addEventListener('mousedown', (e) => {
                    let scrubberX = mousePositionToScrubber(getMousePosition(e));
                    setScrubberPosition(scrubberX);
                    setProgressPlayedWidth(scrubberX);
                    scrubberStartMove(e);
                });

                (<any>this).$refs.progressBar.addEventListener('touchstart', (e) => {
                    let scrubberX = mousePositionToScrubber(getMousePosition(e));
                    setScrubberPosition(scrubberX);
                    setProgressPlayedWidth(scrubberX);
                    scrubberStartMove(e);
                });

                (<any>this).$refs.progressBar.addEventListener('mousemove', (e) => {
                    let scrubberX = mousePositionToScrubber(getMousePosition(e));
                    setProgressHoverWidth(scrubberX);
                });

                (<any>this).$refs.progressBar.addEventListener('mouseout', (e) => {
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

    .progress-bar-scrubber {
        opacity: 0;
        width: $progress-bar-scrubber-size;
        height: $progress-bar-scrubber-size;
        border-radius: 50%;
        background-color: red;
        position: absolute;
        bottom: 0px + $progress-bar-active-height/2 - $progress-bar-scrubber-size/2;
        cursor: pointer;
    }

    .progress-bar {
        top: 0px;
        height: $progress-bar-height;
        //transition: height 0.1s;
        position: absolute;
        bottom: 0px;
        background-color: $progress-bar-color;
        width: 100%;
        cursor: pointer;
    }

    .progress-bar-hover {
        height: 100%;
        position: absolute;
        width: 0px;
        background-color: $progress-bar-hover-color;
    }

    .progress-bar-played {
        height: 100%;
        position: absolute;
        width: 0px;
        background-color: $progress-bar-played-color;
    }
}


.progress-bar-container:hover {

    .progress-bar {
        height: $progress-bar-active-height;
        top: -( $progress-bar-active-height -  $progress-bar-height);
    }
    .progress-bar {
        .progress-bar-scrubber {
            display: block;
            opacity: 1;
        }
    }

}



</style>