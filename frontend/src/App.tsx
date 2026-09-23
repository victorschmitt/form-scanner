import { useState, type ChangeEvent } from 'react'
import './App.css'

const apiUrl = import.meta.env.VITE_API_URL?.trim() ?? ''

type Box = {
  x: number
  y: number
  width: number
  height: number
}

type Field = {
  name: string
  value: string
  boxes: Box[]
}

type ScanResult = {
  width: number
  height: number
  deskewed: boolean
  image: string
  fields: Field[]
  error?: string
}

type Mode = 'upload' | 'camera'

function App() {
  const [mode, setMode] = useState<Mode>('upload')
  const [file, setFile] = useState<File | null>(null)
  const [result, setResult] = useState<ScanResult | null>(null)
  const [status, setStatus] = useState<'idle' | 'sending' | 'error'>('idle')
  const [message, setMessage] = useState('')

  function switchMode(next: Mode) {
    setMode(next)
    setFile(null)
    setResult(null)
    setStatus('idle')
    setMessage('')
  }

  function onPick(event: ChangeEvent<HTMLInputElement>) {
    const next = event.target.files?.[0]
    if (!next) return
    setFile(next)
    setResult(null)
    setStatus('idle')
    setMessage('')
  }

  async function send() {
    if (!file || !apiUrl) return
    setStatus('sending')
    setMessage('')
    setResult(null)

    const body = new FormData()
    body.append('image', file)

    try {
      const response = await fetch(apiUrl, { method: 'POST', body })
      const data = (await response.json()) as ScanResult
      if (!response.ok) {
        throw new Error(data.error ?? `HTTP ${response.status}`)
      }
      setResult(data)
      setStatus('idle')
    } catch (error) {
      setStatus('error')
      setMessage(error instanceof Error ? error.message : 'Envoi impossible.')
    }
  }

  return (
    <main>
      <header>
        <h1>Form Scanner</h1>
        <p className="hint">
          Envoyez un formulaire à l’API depuis un fichier ou depuis l’appareil
          photo, puis inspectez les champs extraits.
        </p>
      </header>

      <div className="modes" role="group" aria-label="Source de l’image">
        <button
          type="button"
          className={mode === 'upload' ? 'mode active' : 'mode'}
          onClick={() => switchMode('upload')}
        >
          Fichier
        </button>
        <button
          type="button"
          className={mode === 'camera' ? 'mode active' : 'mode'}
          onClick={() => switchMode('camera')}
        >
          Photo
        </button>
      </div>

      {!apiUrl && (
        <p className="banner" role="status">
          Configurez <code>VITE_API_URL</code> (ex.{' '}
          <code>http://localhost:8080/scan</code>).
        </p>
      )}

      <div className="actions">
        <label className="pick">
          {mode === 'camera' ? 'Prendre une photo' : 'Choisir une image'}
          <input
            type="file"
            accept="image/*"
            capture={mode === 'camera' ? 'environment' : undefined}
            onChange={onPick}
          />
        </label>
        <button
          type="button"
          disabled={!file || !apiUrl || status === 'sending'}
          onClick={send}
        >
          {status === 'sending' ? 'Analyse…' : 'Envoyer'}
        </button>
        {file && <span className="filename">{file.name}</span>}
      </div>

      {message && (
        <p className="error" role="status">
          {message}
        </p>
      )}

      {result && (
        <section className="result">
          <div className="stage">
            <img
              src={`data:image/jpeg;base64,${result.image}`}
              alt="Formulaire redressé"
            />
            {(result.fields ?? []).flatMap((field) =>
              field.boxes.map((box, i) => (
                <span
                  key={`${field.name}-${i}`}
                  className="box"
                  title={`${field.name}: ${field.value}`}
                  style={{
                    left: `${(box.x / result.width) * 100}%`,
                    top: `${(box.y / result.height) * 100}%`,
                    width: `${(box.width / result.width) * 100}%`,
                    height: `${(box.height / result.height) * 100}%`,
                  }}
                />
              )),
            )}
          </div>
          <aside>
            <p>
              {result.deskewed ? 'Deskew appliqué' : 'Pas de deskew'} ·{' '}
              {result.width}×{result.height} · {result.fields?.length ?? 0}{' '}
              champs
            </p>
            <dl>
              {(result.fields ?? []).map((field) => (
                <div key={field.name}>
                  <dt>{field.name}</dt>
                  <dd>{field.value}</dd>
                </div>
              ))}
            </dl>
          </aside>
        </section>
      )}
    </main>
  )
}

export default App
