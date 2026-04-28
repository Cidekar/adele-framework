<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { apiGet, apiPost } from '../api'

const route = useRoute()
const router = useRouter()
const password = ref('')
const verifyPassword = ref('')
const errors = ref<Record<string, string>>({})
const message = ref('')
const submitting = ref(false)
const tokenValid = ref(true)

const tokenEmail = (): string => {
    const v = route.query.email
    return typeof v === 'string' ? v : ''
}

onMounted(async () => {
    const e = tokenEmail()
    if (!e) {
        tokenValid.value = false
        message.value = 'Missing or invalid reset token.'
        return
    }
    const res = await apiGet(`/reset-password?email=${encodeURIComponent(e)}`)
    if (res.ok === false) {
        tokenValid.value = false
        if (res.message) message.value = res.message
    }
})

async function submit() {
    submitting.value = true
    errors.value = {}
    message.value = ''
    const res = await apiPost('/reset-password', {
        email: tokenEmail(),
        password: password.value,
        'verify-password': verifyPassword.value,
    })
    submitting.value = false
    if (res.ok) {
        router.push(res.redirect ?? '/login')
        return
    }
    if (res.errors) errors.value = res.errors
    if (res.message) message.value = res.message
}
</script>

<template>
    <div class="bg-pink-50">
        <div class="centered-vh">
            <div class="card">
                <div class="header">
                    <h1>Reset Password</h1>
                </div>
                <div class="body">
                    <form @submit.prevent="submit" novalidate @keydown.enter.prevent>
                        <input type="hidden" name="email" :value="tokenEmail()" />

                        <div class="mx-10">
                            <div class="my-4" :class="{ error: errors.password }">
                                <label for="password" class="block text-pink-50">Password</label>
                                <div class="relative">
                                    <input id="password" v-model="password" name="password" class="w-full" type="password" :disabled="!tokenValid" />
                                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor" class=" h-[20px] fill-pink-1000 absolute top-1/2 -translate-y-1/2 right-2">
                                        <path fill-rule="evenodd" d="M8 1a3.5 3.5 0 0 0-3.5 3.5V7A1.5 1.5 0 0 0 3 8.5v5A1.5 1.5 0 0 0 4.5 15h7a1.5 1.5 0 0 0 1.5-1.5v-5A1.5 1.5 0 0 0 11.5 7V4.5A3.5 3.5 0 0 0 8 1Zm2 6V4.5a2 2 0 1 0-4 0V7h4Z" clip-rule="evenodd" />
                                    </svg>
                                </div>
                                <span v-if="errors.password" class="validator-message">{{ errors.password }}</span>
                            </div>

                            <div class="my-4" :class="{ error: errors['verify-password'] }">
                                <label for="verify-password" class="block text-pink-50">Confirm Password</label>
                                <div class="relative">
                                    <input id="verify-password" v-model="verifyPassword" name="verify-password" class="w-full" type="password" :disabled="!tokenValid" />
                                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor" class=" h-[20px] fill-pink-1000 absolute top-1/2 -translate-y-1/2 right-2">
                                        <path fill-rule="evenodd" d="M8 1a3.5 3.5 0 0 0-3.5 3.5V7A1.5 1.5 0 0 0 3 8.5v5A1.5 1.5 0 0 0 4.5 15h7a1.5 1.5 0 0 0 1.5-1.5v-5A1.5 1.5 0 0 0 11.5 7V4.5A3.5 3.5 0 0 0 8 1Zm2 6V4.5a2 2 0 1 0-4 0V7h4Z" clip-rule="evenodd" />
                                    </svg>
                                </div>
                                <span v-if="errors['verify-password']" class="validator-message">{{ errors['verify-password'] }}</span>
                            </div>

                            <div class="my-6 text-center">
                                <input class="button" type="submit" :value="submitting ? 'Submitting...' : 'Submit'" :disabled="submitting || !tokenValid" />
                            </div>
                        </div>
                    </form>
                </div>

                <div class="footer">

                </div>
            </div>
        </div>
    </div>
</template>
