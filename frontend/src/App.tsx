import { useEffect, useState, type ChangeEvent } from 'react'
import './App.css'

const apiUrl = import.meta.env.VITE_API_URL?.trim() ?? ''

function App() {
  const [file, setFile] = useState<File | null>(null)
  const [preview, setPreview] = useState<string | null>(null)
  const [status, setStatus] = useState<'idle' | 'sending' | 'ok' | 'error'>(
    'idle',
  )
  const [message, setMessage] = useState('')

  useEffect(() => {
    return () => {
      if (preview) URL.revokeObjectURL(preview)
    }
  }, [preview])

  function onCapture(event: ChangeEvent<HTMLInputElement>) {
    const next = event.target.files?.[0]
    if (!next) return
    setFile(next)
    setPreview((current) => {
      if (current) URL.revokeObjectURL(current)
      return URL.createObjectURL(next)
    })
    setStatus('idle')
    setMessage('')
  }

  async function send() {
    if (!file || !apiUrl) return
    setStatus('sending')
    setMessage('')

    const body = new FormData()
    body.append('image', file)

    try {
      const response = await fetch(apiUrl, { method: 'POST', body })
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`)
      }
      setStatus('ok')
      setMessage('Image envoyée.')
    } catch (error) {
      setStatus('error')
      setMessage(error instanceof Error ? error.message : 'Envoi impossible.')
    }
  }

  return (
    <main>
      <h1>Form Scanner</h1>
      <p className="hint">Prenez une photo, puis envoyez-la à l’API.</p>

      {!apiUrl && (
        <p className="banner" role="status">
          Configurez <code>VITE_API_URL</code> (fichier <code>.env</code> en
          local, variables d’environnement sur Vercel).
        </p>
      )}

      <label className="capture">
        Prendre une photo
        <input
          type="file"
          accept="image/*"
          capture="environment"
          onChange={onCapture}
        />
      </label>

      {preview && (
        <img className="preview" src={preview} alt="Aperçu de la photo" />
      )}

      <button type="button" disabled={!file || !apiUrl || status === 'sending'} onClick={send}>
        {status === 'sending' ? 'Envoi…' : 'Envoyer'}
      </button>

      {message && (
        <p className={status === 'error' ? 'error' : 'ok'} role="status">
          {message}
        </p>
      )}
    </main>
  )
}

export default App
