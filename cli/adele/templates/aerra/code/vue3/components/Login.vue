<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { apiPost } from '../api'

const router = useRouter()
const email = ref('')
const password = ref('')
const errors = ref<Record<string, string>>({})
const message = ref('')
const submitting = ref(false)

async function submit() {
    submitting.value = true
    errors.value = {}
    message.value = ''
    const res = await apiPost('/login', {
        email: email.value,
        password: password.value,
    })
    submitting.value = false
    if (res.ok) {
        router.push(res.redirect ?? '/dashboard/home')
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
                    <h1>Login</h1>
                </div>

                <div class="body">
                    <form @submit.prevent="submit" novalidate @keydown.enter.prevent>
                        <div class="mx-10">
                            <div class="my-4" :class="{ error: errors.email }">
                                <label for="email" class="block text-pink-50">Email</label>
                                <div class="relative">
                                    <input id="email" v-model="email" name="email" class="w-full" type="text" placeholder="aerra@adele.com" />
                                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor" class=" h-[16px] h-[19px] fill-pink-1000 absolute top-1/2 -translate-y-1/2 right-2">
                                        <path d="M2.5 3A1.5 1.5 0 0 0 1 4.5v.793c.026.009.051.02.076.032L7.674 8.51c.206.1.446.1.652 0l6.598-3.185A.755.755 0 0 1 15 5.293V4.5A1.5 1.5 0 0 0 13.5 3h-11Z" />
                                        <path d="M15 6.954 8.978 9.86a2.25 2.25 0 0 1-1.956 0L1 6.954V11.5A1.5 1.5 0 0 0 2.5 13h11a1.5 1.5 0 0 0 1.5-1.5V6.954Z" />
                                    </svg>
                                </div>
                                <span v-if="errors.email" class="validator-message">{{ errors.email }}</span>
                            </div>

                            <div class="my-4" :class="{ error: errors.password }">
                                <label for="password" class="block text-pink-50">Password</label>
                                <div class="relative">
                                    <input id="password" v-model="password" name="password" class="w-full" type="password" />
                                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor" class=" h-[20px] fill-pink-1000 absolute top-1/2 -translate-y-1/2 right-2">
                                        <path fill-rule="evenodd" d="M8 1a3.5 3.5 0 0 0-3.5 3.5V7A1.5 1.5 0 0 0 3 8.5v5A1.5 1.5 0 0 0 4.5 15h7a1.5 1.5 0 0 0 1.5-1.5v-5A1.5 1.5 0 0 0 11.5 7V4.5A3.5 3.5 0 0 0 8 1Zm2 6V4.5a2 2 0 1 0-4 0V7h4Z" clip-rule="evenodd" />
                                    </svg>
                                </div>
                                <span v-if="errors.password" class="validator-message">{{ errors.password }}</span>
                            </div>

                            <div class="my-6 text-center">
                                <input class="button" type="submit" :value="submitting ? 'Submitting...' : 'login'" :disabled="submitting" />
                            </div>
                        </div>
                    </form>
                </div>

                <div class="footer">
                    <p>I want an account, <RouterLink to="/registration">sign up</RouterLink>.</p>
                    <p><RouterLink class="underline ml-2" to="/forgot">Forgot password</RouterLink></p>
                </div>
            </div>
        </div>
    </div>
</template>
