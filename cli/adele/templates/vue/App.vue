<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref } from 'vue'

interface Phrase {
    value: string
}

const phrases: Phrase[] = [
    { value: 'Hey.' },
    { value: 'You can skip the wiring code.' },
    { value: 'Batteries are included.' },
    { value: 'ADELE' },
]

const consoleText = ref('')
const cursorVisible = ref(true)

let phraseIndex = 0
let letterCount = 0
let direction: 'ltr' | 'rtl' = 'ltr'
let typeTimer: number | null = null
let cursorTimer: number | null = null

const speeds = {
    flash: 400,
    typeSlow: 120,
    typeFast: 70,
    delay: 1000,
}

function tick() {
    const current = phrases[phraseIndex].value
    consoleText.value = current.substring(0, letterCount)

    if (direction === 'ltr') {
        letterCount += 1
        if (letterCount > current.length) {
            // Pause at the full word, then start deleting.
            window.clearInterval(typeTimer as number)
            window.setTimeout(() => {
                direction = 'rtl'
                letterCount = current.length - 1
                schedule()
            }, speeds.delay)
        }
    } else {
        letterCount -= 1
        if (letterCount < 0) {
            // Cycle to next phrase.
            window.clearInterval(typeTimer as number)
            phraseIndex = (phraseIndex + 1) % phrases.length
            direction = 'ltr'
            letterCount = 0
            window.setTimeout(schedule, speeds.delay / 2)
        }
    }
}

function schedule() {
    if (typeTimer !== null) window.clearInterval(typeTimer)
    const speed = direction === 'ltr' ? speeds.typeSlow : speeds.typeFast
    typeTimer = window.setInterval(tick, speed)
}

onMounted(() => {
    cursorTimer = window.setInterval(() => {
        cursorVisible.value = !cursorVisible.value
    }, speeds.flash)
    window.setTimeout(schedule, 1000)
})

onBeforeUnmount(() => {
    if (typeTimer !== null) window.clearInterval(typeTimer)
    if (cursorTimer !== null) window.clearInterval(cursorTimer)
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
