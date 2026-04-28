<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref } from 'vue'
import { RouterLink } from 'vue-router'

const phrases = ['Hey.', 'You can skip the wiring code.', 'Batteries are included.', 'ADELE']
const consoleText = ref('')
const cursorVisible = ref(true)

let phraseIndex = 0
let letterCount = 0
let direction: 'ltr' | 'rtl' = 'ltr'
let typeTimer: number | null = null
let cursorTimer: number | null = null

const speeds = { flash: 400, typeSlow: 120, typeFast: 70, delay: 1000 }

function tick() {
    const current = phrases[phraseIndex]
    consoleText.value = current.substring(0, letterCount)
    if (direction === 'ltr') {
        letterCount += 1
        if (letterCount > current.length) {
            if (typeTimer !== null) window.clearInterval(typeTimer)
            window.setTimeout(() => { direction = 'rtl'; letterCount = current.length - 1; schedule() }, speeds.delay)
        }
    } else {
        letterCount -= 1
        if (letterCount < 0) {
            if (typeTimer !== null) window.clearInterval(typeTimer)
            phraseIndex = (phraseIndex + 1) % phrases.length
            direction = 'ltr'; letterCount = 0
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
    cursorTimer = window.setInterval(() => { cursorVisible.value = !cursorVisible.value }, speeds.flash)
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
            <div class="text-pink-100 text-sm not-italic font-light md:mt-4">
                <RouterLink class="mx-1" to="/login">Login</RouterLink>
                <RouterLink class="mx-1" to="/registration">Registration</RouterLink>
                <RouterLink class="mx-1" to="/forgot">Forgot Password</RouterLink>
            </div>
            <div class="not-italic mt-6 flex justify-center">
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 261.76 226.69" class="h-[40px] w-auto" aria-label="Vue">
                    <path d="M161.096.001l-30.225 52.351L100.647.001H-.005l130.877 226.688L261.749.001z" fill="#41b883"/>
                    <path d="M161.096.001l-30.225 52.351L100.647.001H52.346l78.526 136.01L209.398.001z" fill="#34495e"/>
                </svg>
            </div>
        </div>
    </div>
</template>
