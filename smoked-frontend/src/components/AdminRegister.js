import React, { useState } from 'react';
import api from '../utils/api';
import { useNavigate } from 'react-router-dom';

function AdminRegister() {
    const [formData, setFormData] = useState({
        username: '',
        password: '',
        phone_number: '',
        name: '',
        role: 'client',
    });
    const [error, setError] = useState('');
    const [success, setSuccess] = useState('');
      const navigate = useNavigate();

    const handleChange = (e) => {
        setFormData({ ...formData, [e.target.name]: e.target.value });
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        setError('');
        setSuccess('');
        try {
            const token = localStorage.getItem('token');
            const response = await api.post('/admin/register', formData, {
                headers: { Authorization: `Bearer ${token}` },
            });
            setSuccess(response.data.message);
            setFormData({ username: '', password: '', phone_number: '', name: '', role: 'client' });
        } catch (err) {
            setError(err.response?.data?.error || 'Registration failed');
        }
    };

     const Back = () => {
    navigate('/');
  };

    return (
        <div className="container mt-4">
            <h2>Register New User (Admin)</h2>
            {error && <div className="alert alert-danger">{error}</div>}
            {success && <div className="alert alert-success">{success}</div>}
            <form onSubmit={handleSubmit}>
                <div className="mb-3">
                    <label htmlFor="username" className="form-label">Username</label>
                    <input
                        type="text"
                        className="form-control"
                        id="username"
                        name="username"
                        value={formData.username}
                        onChange={handleChange}
                        required
                        minLength="3"
                        maxLength="50"
                    />
                </div>
                <div className="mb-3">
                    <label htmlFor="password" className="form-label">Password</label>
                    <input
                        type="password"
                        className="form-control"
                        id="password"
                        name="password"
                        value={formData.password}
                        onChange={handleChange}
                        required
                        minLength="6"
                    />
                </div>
            <div className="mb-3">
                  <label htmlFor="phone_number" className="form-label">Phone number:</label>
                  <input 
                    type="tel" 
                    className="form-label"
                    id="phone_number"
                    name="phone_number"
                    value={formData.phone_number} 
                    style={{
                      width: '200px',
                      backgroundColor: 'transparent',
                      color: '#FFFF00',
                      textShadow: '1px 1px 2px rgba(0, 0, 0, 0.8)',
                      marginTop: '5px',
                          }}
                    placeholder="+7 (978) 123-45-67" 
                    onChange={handleChange}
                    required />
                </div>
                <div className="mb-3">
                    <label htmlFor="name" className="form-label">Name (Not necessarily)</label>
                    <input
                        type="text"
                        className="form-control"
                        id="name"
                        name="name"
                        value={formData.name}
                        onChange={handleChange}
                        minLength="2"
                    />
                </div>
                <div className="mb-3">
                    <label htmlFor="role" className="form-label">Role</label>
                    <select
                        className="form-control"
                        id="role"
                        name="role"
                        value={formData.role}
                        onChange={handleChange}
                    >
                        <option value="client">Client</option>
                        <option value="owner">Owner</option>
                    </select>
                </div>
                <button onClick={Back} className="btn btn-primary">Back</button>
                <button type="submit" className="btn btn-primary">Register</button>
            </form>
        </div>
    );
}

export default AdminRegister;