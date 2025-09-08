// pages/_app.js
import '../styles/globals.css'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useState, useEffect } from 'react'
import { useAuthStore } from '../utils/store'

export default function MyApp({ Component, pageProps }) {
    const [qc] = useState(() => new QueryClient())
    const initialize = useAuthStore(state => state.initialize)

    useEffect(() => {
        // Инициализируем состояние аутентификации при загрузке приложения
        initialize()
    }, [initialize])

    return (
        <QueryClientProvider client={qc}>
            <Component {...pageProps} />
        </QueryClientProvider>
    )
}
