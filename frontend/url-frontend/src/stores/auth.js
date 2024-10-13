import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useAuthStore= defineStore('auth', () => {

  const token = ref(null)
  const loading = ref(false)
  const error = ref(null)

  const login = async (username, password) => {
    console.log('login', username, password)

    try {
      const response = await fetch(import.meta.env.VITE_AUTH_URL, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ username, password })
      });

      console.log(response)
      if (!response.ok) {
        throw new Error('Invalid credentials');
      }

      const data = await response.json();
      console.log(data)
      console.log(data.token)
      token.value = data.token;
    } catch (err) {
      error.value = err;
    } finally {
      loading.value = false;

    }
  };

  const logout = () => {
    token.value = null;
  }

  return { token, loading, error, login, logout }
}, {
  persist: true
});