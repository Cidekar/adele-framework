<script lang="ts">
import Vue from 'vue'

interface Phrase {
    value: string
}

const phrases: Phrase[] = [
    { value: 'Hey.' },
    { value: 'You can skip the wiring code.' },
    { value: 'Batteries are included.' },
    { value: 'ADELE' },
]

const speeds = {
    flash: 400,
    typeSlow: 120,
    typeFast: 70,
    delay: 1000,
}

export default Vue.extend({
    data() {
        return {
            consoleText: '',
            cursorVisible: true,
            phraseIndex: 0,
            letterCount: 0,
            direction: 'ltr' as 'ltr' | 'rtl',
            typeTimer: null as number | null,
            cursorTimer: null as number | null,
        }
    },
    mounted() {
        this.cursorTimer = window.setInterval(() => {
            this.cursorVisible = !this.cursorVisible
        }, speeds.flash)
        window.setTimeout(this.schedule, 1000)
    },
    beforeDestroy() {
        if (this.typeTimer !== null) window.clearInterval(this.typeTimer)
        if (this.cursorTimer !== null) window.clearInterval(this.cursorTimer)
    },
    methods: {
        tick(): void {
            const current = phrases[this.phraseIndex].value
            this.consoleText = current.substring(0, this.letterCount)

            if (this.direction === 'ltr') {
                this.letterCount += 1
                if (this.letterCount > current.length) {
                    if (this.typeTimer !== null) window.clearInterval(this.typeTimer)
                    window.setTimeout(() => {
                        this.direction = 'rtl'
                        this.letterCount = current.length - 1
                        this.schedule()
                    }, speeds.delay)
                }
            } else {
                this.letterCount -= 1
                if (this.letterCount < 0) {
                    if (this.typeTimer !== null) window.clearInterval(this.typeTimer)
                    this.phraseIndex = (this.phraseIndex + 1) % phrases.length
                    this.direction = 'ltr'
                    this.letterCount = 0
                    window.setTimeout(this.schedule, speeds.delay / 2)
                }
            }
        },
        schedule(): void {
            if (this.typeTimer !== null) window.clearInterval(this.typeTimer)
            const speed = this.direction === 'ltr' ? speeds.typeSlow : speeds.typeFast
            this.typeTimer = window.setInterval(this.tick, speed)
        },
    },
})
</script>

<template>
    <div class="h-screen flex flex-col items-center justify-center">
        <div class="console before:content-['>'] text-pink-100 h-[200px] italic text-center md:text-[64px] text-[44px] font-bold leading-[60px]">
            <span>{{ consoleText }}</span>
            <span class="not-italic top-[-4px] left-[10px] inline-block" :class="cursorVisible ? '' : 'opacity-0'">_</span>
            <div class="text-pink-100 text-lg font-light mt-2 not-italic">Vue</div>
            <div class="text-pink-100 text-sm not-italic font-light md:mt-4">
                <a class="mx-1" href="/login">Login</a>
                <a class="mx-1" href="/registration">Registration</a>
                <a class="mx-1" href="/forgot">Forgot Password</a>
                <a class="mx-1" href="/404">404 Demo</a>
            </div>
        </div>
    </div>
</template>
