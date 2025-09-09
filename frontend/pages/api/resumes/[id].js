// pages/api/resumes/[id].js
export default async function handler(req, res) {
    const { id } = req.query;
    const token = req.headers.authorization;

    if (!token) {
        return res.status(401).json({ error: 'Unauthorized' });
    }

    if (req.method === 'DELETE') {
        try {
            const response = await fetch(`http://localhost:8081/api/resumes/${id}`, {
                method: 'DELETE',
                headers: {
                    'Authorization': token,
                },
            });

            if (response.ok) {
                const data = await response.json();
                res.json(data);
            } else {
                const errorData = await response.json();
                res.status(response.status).json(errorData);
            }
        } catch (error) {
            console.error('API Error:', error);
            res.status(500).json({ error: 'Server error' });
        }
    } else if (req.method === 'GET') {
        try {
            const response = await fetch(`http://localhost:8081/api/resumes/${id}`, {
                headers: {
                    'Authorization': token,
                },
            });

            const data = await response.json();

            if (response.ok) {
                res.json(data);
            } else {
                res.status(response.status).json(data);
            }
        } catch (error) {
            console.error('API Error:', error);
            res.status(500).json({ error: 'Server error' });
        }
    } else {
        res.setHeader('Allow', ['GET', 'DELETE']);
        res.status(405).json({ error: 'Method not allowed' });
    }
}
