import { useState } from 'react'
import DeveloperDashboard from './components/DeveloperDashboard'
import UserDashboard from './components/UserDashboard'

function App() {
  const [activeTab, setActiveTab] = useState('developer')
  const [orgId, setOrgId] = useState('22222222-2222-2222-2222-222222222222') // Default to Dev Org
  const [jwtToken, setJwtToken] = useState(localStorage.getItem('jwtToken') || '')

  const handleTokenChange = (e) => {
    const token = e.target.value
    setJwtToken(token)
    localStorage.setItem('jwtToken', token)
  }

  return (
    <div className="container">
      <h1>🎯 Stripe Connect Marketplace - Testing Dashboard</h1>
      <p style={{ color: '#666', marginBottom: '20px' }}>
        Test the complete Stripe Connect Express integration for your Function Marketplace
      </p>

      <div className="card" style={{ marginBottom: '20px', backgroundColor: '#fff3cd', borderColor: '#ffc107' }}>
        <h3 style={{ marginTop: 0, color: '#856404' }}>⚠️ Authentication Required</h3>
        <div className="form-group" style={{ marginBottom: '10px' }}>
          <label htmlFor="jwtToken" style={{ fontWeight: 'bold' }}>JWT Token:</label>
          <input
            id="jwtToken"
            type="password"
            value={jwtToken}
            onChange={handleTokenChange}
            placeholder="Paste your JWT Bearer token here"
            style={{ fontFamily: 'monospace', fontSize: '12px' }}
          />
          <small style={{ color: '#856404', fontSize: '12px', display: 'block', marginTop: '5px' }}>
            Get your JWT token from the Rival API login endpoint. This token is stored in localStorage.
          </small>
        </div>
        {!jwtToken && (
          <div style={{ padding: '10px', backgroundColor: '#f8d7da', borderRadius: '4px', color: '#721c24', fontSize: '14px' }}>
            ❌ <strong>Missing JWT Token:</strong> You must provide a valid JWT token to use this dashboard.
          </div>
        )}
      </div>

      <div className="org-selector">
        <label htmlFor="orgId">Organization ID (for testing):</label>
        <input
          id="orgId"
          type="text"
          value={orgId}
          onChange={(e) => setOrgId(e.target.value)}
          placeholder="Enter your organization ID"
        />
      </div>

      <div className="tabs">
        <button
          className={`tab ${activeTab === 'developer' ? 'active' : ''}`}
          onClick={() => setActiveTab('developer')}
        >
          👨‍💻 Developer Dashboard
        </button>
        <button
          className={`tab ${activeTab === 'user' ? 'active' : ''}`}
          onClick={() => setActiveTab('user')}
        >
          👤 User Dashboard
        </button>
      </div>

      {activeTab === 'developer' && <DeveloperDashboard orgId={orgId} jwtToken={jwtToken} />}
      {activeTab === 'user' && <UserDashboard orgId={orgId} jwtToken={jwtToken} />}
    </div>
  )
}

export default App
