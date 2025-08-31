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
    const { isAuthenticated, logout } = useContext(AuthContext);
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
                       {isAuthenticated ? (
                           <>
                               <Link to="/register" className="btn btn-sm btn-warning me-2">Register new user</Link>
                               <button onClick={logout} className="btn btn-link btn-danger">Logout</button>
                           </>
                       ) : (
                           <>
                               <Link to="/login" className="me-2">Login</Link>
                              {/* SSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSS */}
                           </>
                       )}
                   </nav>
                   <Routes>
                       <Route path="/" element={<AssortmentList isAuthenticated={isAuthenticated} />} />
                       <Route path="/add" element={isAuthenticated ? <ProductForm /> : <Navigate to="/login" />} />
                       <Route path="/edit/:id" element={isAuthenticated ? <ProductForm /> : <Navigate to="/login" />} />
                       <Route path="/calculate" element={<CalculatePrice />} />
                       <Route path="/login" element={<Login />} />
                       <Route path="/register" element={<Register />} />
                       <Route path="/cart" element={<Cart />} />
                   </Routes>
               </div>
           </div>
       );
   }

   export default App;