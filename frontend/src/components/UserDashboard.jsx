import { useState, useEffect } from 'react'
import axios from 'axios'

function UserDashboard({ orgId, jwtToken }) {
  const [functionId, setFunctionId] = useState('func-test-123')
  const [version, setVersion] = useState('v1.0.0')
  const [developerOrgId, setDeveloperOrgId] = useState('22222222-2222-2222-2222-222222222222')
  const [amount, setAmount] = useState(5.00)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [success, setSuccess] = useState(null)
  const [lastTransaction, setLastTransaction] = useState(null)
  const [connectedDevelopers, setConnectedDevelopers] = useState([])
  const [loadingDevelopers, setLoadingDevelopers] = useState(false)

  const API_URL = 'http://localhost:8080/api/v1/connect'

  useEffect(() => {
    if (jwtToken) {
      loadConnectedDevelopers()
    }
  }, [orgId, jwtToken])

  const getHeaders = () => ({
    'X-Organization-ID': orgId,
    'Authorization': `Bearer ${jwtToken}`,
  })

  const loadConnectedDevelopers = async () => {
    if (!jwtToken) return

    setLoadingDevelopers(true)
    try {
      const headers = getHeaders()
      const response = await axios.get(`${API_URL}/connected-developers`, { headers })
      setConnectedDevelopers(response.data.developers || [])
    } catch (err) {
      console.error('Failed to load connected developers:', err)
    } finally {
      setLoadingDevelopers(false)
    }
  }

  const handleExecuteFunction = async () => {
    if (!jwtToken) {
      setError('JWT token is required')
      return
    }

    setError(null)
    setSuccess(null)
    setLoading(true)
    setLastTransaction(null)

    try {
      const headers = getHeaders()
      const response = await axios.post(`${API_URL}/payments/execute`, {
        function_id: functionId,
        version: version,
        amount: amount,
        developer_organization_id: developerOrgId,
      }, { headers })

      setLastTransaction(response.data)
      setSuccess(`Payment successful! Transaction ID: ${response.data.transaction_id}`)
      loadConnectedDevelopers() // Reload connected developers
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to process payment')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div>
      <div className="card">
        <h2>Execute Function & Pay Developer</h2>
        <p style={{ color: '#666', marginBottom: '20px' }}>
          When you execute a function, the amount is deducted from your wallet balance and
          transferred to the developer's wallet.
        </p>

        {error && <div className="alert alert-error">{error}</div>}
        {success && <div className="alert alert-success">{success}</div>}

        <div className="form-group">
          <label htmlFor="functionId">Function ID</label>
          <input
            id="functionId"
            type="text"
            value={functionId}
            onChange={(e) => setFunctionId(e.target.value)}
            placeholder="Enter function ID"
          />
          <small style={{ color: '#666', fontSize: '12px' }}>
            This is a test function ID. In production, this would come from your function catalog.
          </small>
        </div>

        <div className="form-group">
          <label htmlFor="version">Function Version</label>
          <input
            id="version"
            type="text"
            value={version}
            onChange={(e) => setVersion(e.target.value)}
            placeholder="e.g., v1.0.0"
          />
          <small style={{ color: '#666', fontSize: '12px' }}>
            The version of the function to execute. Default: v1.0.0
          </small>
        </div>

        <div className="form-group">
          <label htmlFor="developerOrgId">Developer Organization ID</label>
          <input
            id="developerOrgId"
            type="text"
            value={developerOrgId}
            onChange={(e) => setDeveloperOrgId(e.target.value)}
            placeholder="Enter developer organization ID"
          />
          <small style={{ color: '#666', fontSize: '12px' }}>
            The organization ID of the developer who owns the function. Default: 22222222-2222-2222-2222-222222222222
          </small>
        </div>

        <div className="form-group">
          <label htmlFor="amount">Amount (USD)</label>
          <input
            id="amount"
            type="number"
            min="0.01"
            step="0.01"
            value={amount}
            onChange={(e) => setAmount(parseFloat(e.target.value))}
          />
          <small style={{ color: '#666', fontSize: '12px' }}>
            This would typically be the function's price_per_api_request from your database.
          </small>
        </div>

        <button
          className="btn btn-primary"
          onClick={handleExecuteFunction}
          disabled={loading || !functionId || amount <= 0}
        >
          {loading ? 'Processing...' : `Execute Function & Pay $${amount.toFixed(2)}`}
        </button>
      </div>

      {lastTransaction && (
        <div className="card">
          <h2>Transaction Details</h2>
          <div className="info-grid">
            <div className="info-item">
              <label>Transaction ID</label>
              <div className="value" style={{ fontSize: '14px' }}>
                {lastTransaction.transaction_id}
              </div>
            </div>
            <div className="info-item">
              <label>Amount Paid</label>
              <div className="value">${lastTransaction.amount.toFixed(2)}</div>
            </div>
            <div className="info-item">
              <label>Platform Fee</label>
              <div className="value">${lastTransaction.platform_fee.toFixed(2)}</div>
            </div>
            <div className="info-item">
              <label>Developer Receives</label>
              <div className="value success">${lastTransaction.net_amount.toFixed(2)}</div>
            </div>
            <div className="info-item">
              <label>Your New Balance</label>
              <div className="value">${lastTransaction.user_balance.toFixed(2)}</div>
            </div>
            <div className="info-item">
              <label>Developer Balance</label>
              <div className="value">${lastTransaction.developer_balance.toFixed(2)}</div>
            </div>
          </div>
        </div>
      )}

      <div className="card">
        <h2>Connected Developers</h2>
        <p style={{ color: '#666', marginBottom: '15px' }}>
          Developers you have paid for function executions
        </p>

        {loadingDevelopers ? (
          <div style={{ padding: '20px', textAlign: 'center', color: '#666' }}>
            Loading...
          </div>
        ) : connectedDevelopers.length === 0 ? (
          <div style={{ padding: '20px', textAlign: 'center', color: '#666' }}>
            No connected developers yet. Execute a function to see developers here.
          </div>
        ) : (
          <div style={{ overflowX: 'auto' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse' }}>
              <thead>
                <tr style={{ borderBottom: '2px solid #eee', textAlign: 'left' }}>
                  <th style={{ padding: '12px 8px' }}>Developer Org ID</th>
                  <th style={{ padding: '12px 8px' }}>Status</th>
                  <th style={{ padding: '12px 8px' }}>Balance</th>
                  <th style={{ padding: '12px 8px' }}>Total Earned</th>
                  <th style={{ padding: '12px 8px' }}>Connected Since</th>
                </tr>
              </thead>
              <tbody>
                {connectedDevelopers.map((dev, index) => (
                  <tr key={index} style={{ borderBottom: '1px solid #eee' }}>
                    <td style={{ padding: '12px 8px', fontFamily: 'monospace', fontSize: '12px' }}>
                      {dev.organization_id.substring(0, 20)}...
                    </td>
                    <td style={{ padding: '12px 8px' }}>
                      {dev.onboarding_completed ? (
                        <span style={{ color: '#10b981', fontSize: '14px' }}>✓ Onboarded</span>
                      ) : (
                        <span style={{ color: '#f59e0b', fontSize: '14px' }}>⚠ Pending</span>
                      )}
                    </td>
                    <td style={{ padding: '12px 8px', fontWeight: 'bold' }}>
                      ${dev.balance.toFixed(2)}
                    </td>
                    <td style={{ padding: '12px 8px', color: '#10b981' }}>
                      ${dev.total_earned.toFixed(2)}
                    </td>
                    <td style={{ padding: '12px 8px', color: '#666', fontSize: '14px' }}>
                      {dev.joined_at}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      <div className="card">
        <h2>How It Works</h2>
        <ol style={{ paddingLeft: '20px', lineHeight: '1.8' }}>
          <li><strong>User Tops Up Wallet</strong> - Users add funds to their account balance via Stripe Payment Intent (existing functionality)</li>
          <li><strong>Execute Function</strong> - User executes a function and the price is deducted from their balance</li>
          <li><strong>Developer Gets Paid</strong> - The amount (minus platform fee) is added to the developer's wallet</li>
          <li><strong>Developer Withdraws</strong> - When balance reaches $50+, developer can withdraw to their bank account via Stripe Connect</li>
        </ol>
      </div>

      <div className="card">
        <h2>Testing Notes</h2>
        <div className="alert alert-info">
          <strong>Before testing function execution:</strong>
          <ol style={{ paddingLeft: '20px', marginTop: '10px' }}>
            <li>Make sure you have a user account with sufficient balance in the database</li>
            <li>The developer organization must exist in the database</li>
            <li>In production, the function_id would reference an actual function in your functions table</li>
            <li>The developer organization ID is currently hardcoded as "placeholder_developer_org_id" in the service - update this to match your test developer</li>
          </ol>
        </div>
      </div>
    </div>
  )
}

export default UserDashboard
