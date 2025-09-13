import React, { useState, useEffect } from 'react';
import api from '../utils/api';
import { Link } from 'react-router-dom';
import {jwtDecode }from 'jwt-decode';
import './AssortmentList.css';
import ReactPaginate from 'react-paginate'; 

const ClientOrders = ({ isAuthenticated }) => {
    const [orders, setOrders] = useState([]);
    const [error, setError] = useState(null);
    const [loading, setLoading] = useState(true); // Состояние загрузки, если данные не успели попасть в рендер
    const [page, setPage] = useState(1); 
    const [size] = useState(10); 
    const [totalPages, setTotalPages] = useState(1); 
    const [totalCount, setTotalCount] = useState(0); 

    useEffect(() => {
        fetchMyOrders(page, size);
      }, [page, size]);

   const fetchMyOrders = async (page, size) => {
    try {
      const response = await api.get('/client/orders', {
        params: { page, size },
      });
      setOrders(response.data.orders || []);
      setTotalPages(response.data.total_pages || 1);
      setTotalCount(response.data.total_count || 0);
      setError('');
      setLoading(false);
    } catch (error) {
      console.error('Ошибка при загрузке заказов:', error);
      setError('Не удалось загрузить заказы');
    }
  };

  const handlePageChange = ({ selected }) => {
    setPage(selected + 1); // react-paginate использует 0-based индексы
  };

    if (error) {
        return <div>{error}</div>;
    }

    if (loading) {
        return <div>Loading...</div>; 
    }

    if (orders.length === 0 && !error && totalCount === 0) {
    return <div>No orders found.</div>;
    }

    return (
           <div>
                <Link to="/" className="btn btn-warning mb-3">
                    Back to Assortment
                </Link>
                        {isAuthenticated && Array.isArray(orders) && orders.length > 0 && (  
                            <>  
                            <table className="table table-striped table-transparent">
                                <thead>
                                <tr>
                                    <th style={{textShadow: '2px 2px 4px rgba(255, 0, 0, 0.8)' }}>Created Time</th>
                                    <th>Products</th>
                                    <th>Total Price</th>
                                    <th>Phone Number</th>
                                    <th>Name</th>
                                </tr>
                            </thead>
                            <tbody>
                                {orders.map((order) => (
                                    <tr key={order.id}>               
                                        <td>{order.created_at}</td>
                                        
                        <td>{order.items.map((item) => (
                            <div>
                                    <p style={{textShadow: '2px 2px 4px rgba(255, 0, 0, 0.8)' }}>{item.meat}, {item.quantity}</p>
                                    <p>{item.selected_spice}</p>
                            </div>
                        
                        ))}</td>
                                        <td>{order.total_price} rub.</td>
                                        <td>{order.phone_number}</td>
                                        <td>{order.name}</td>
                                    </tr>
                                ))}
                            </tbody>
                            </table>
                            <ReactPaginate
            previousLabel="Previous"
            nextLabel="Next"
            pageCount={totalPages}
            onPageChange={handlePageChange}
            containerClassName="pagination"
            pageClassName="page-item"
            pageLinkClassName="page-link"
            previousClassName="page-item"
            nextClassName="page-item"
            previousLinkClassName="page-link"
            nextLinkClassName="page-link"
            activeClassName="active"
            disabledClassName="disabled"
          />
          </>
                        )}
                    </div>
    );
}

export default ClientOrders;