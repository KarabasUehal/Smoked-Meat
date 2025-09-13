import React, { useContext } from 'react';
import { BrowserRouter as Router, Routes, Route, Link, Navigate } from 'react-router-dom';
import AssortmentList from './components/AssortmentList';
import ProductForm from './components/ProductForm';
import CalculatePrice from './components/CalculatePrice';
import Login from './components/Login';
import Register from './components/Register';
import Cart from './components/Cart';
import { AuthProvider, AuthContext } from './context/AuthContext';
import { CartProvider } from './context/CartContext';
import backgroundImage from './assets/background.jpg';
import 'bootstrap/dist/css/bootstrap.min.css';
import './App.css';
import OrdersList from './components/OrdersList';
import ClientOrders from './components/ClientOrders';
import AdminRegister from './components/AdminRegister';

   function App() {
    return (
        <AuthProvider>
            <CartProvider>
                <Router>
                    <AppContent />
                </Router>
            </CartProvider>
        </AuthProvider>
    );
}

   function AppContent() {
    const { isAuthenticated, role, logout } = useContext(AuthContext);
    console.log('AppContent: isAuthenticated =', isAuthenticated);

       return (
           <div
               className="app-background" style={{ backgroundImage: `url(${backgroundImage})` }}
           >
               <div className="container mt-4">
                   <h1>Smoked meat by Arthur, come and taste!</h1>
                   <nav className="mb-3">
                       <Link to="/" className="btn btn-sm btn-success me-2">Assortment</Link>
                       <Link to="/cart" className="btn btn-sm btn-success me-2">My cart</Link>
                       {isAuthenticated && role ? (
                           <>
                               {role === 'owner' && (
                                <>
                                    <Link to="/orders" className="btn btn-sm btn-success me-2">Orders</Link>
                                    <Link to="/admin/register" className="btn btn-sm btn-success me-2">Register New User</Link>
                                </>
                            )}
                            {role === 'client' && (
                                <Link to="/my-orders" className="btn btn-sm btn-success me-2">My Orders</Link>
                            )}
                            <button onClick={logout} className="btn btn-link btn-danger">Logout</button>
                           </>
                       ) : (
                           <>
                               <Link to="/register" className="btn btn-sm btn-warning me-2">Register</Link>
                               <Link to="/login" className="me-2">Login</Link>
                           </>
                       )}
                   </nav>
                   <Routes>
                       <Route path="/" element={<AssortmentList isAuthenticated={isAuthenticated} />} />
                       <Route path="/add" element={isAuthenticated && role === 'owner'? <ProductForm /> : <Navigate to="/login" />} />
                       <Route path="/edit/:id" element={isAuthenticated && role === 'owner'? <ProductForm /> : <Navigate to="/login" />} />
                       <Route path="/calculate" element={<CalculatePrice />} />
                       <Route path="/login" element={<Login />} />
                       <Route path="/register" element={<Register />} />
                       <Route path="/admin/register" element={(isAuthenticated && role === 'owner') ? <AdminRegister /> : <Navigate to="/login" />}/>
                       <Route path="/cart" element={<Cart isAuthenticated={isAuthenticated}/>} />
                       <Route path="/orders" element={<OrdersList isAuthenticated={isAuthenticated}/>} />
                       <Route path="/my-orders" element={<ClientOrders isAuthenticated={isAuthenticated}/>} />
                   </Routes>
               </div>
           </div>
       );
   }

   export default App;