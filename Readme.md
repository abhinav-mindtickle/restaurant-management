Steps To Run The Code
1. Change the mongo connection string in database/databaseConnection.go 
2. Run mongodb server
3. Change the directory to the RESTAURANT-MANAGEMENT folder
4. In the command line type:
    -> go run main.go

5. Your server will start running on port 8000

6. check routes folder for all the routes

Sample data for post requests:

Usersignup :
    {
        "first_name" : "abhinav",
        "last_name" : "kumar",
        "email" : "abhinav@gmail.com",
        "password" : "1234567",
        "phone" : "23423423"
    }

Login : 
    {
        "email": "abhinav@gmail.com",
        "password" : "1234567"
    }

This will provide the token Id, use it as header named 'token' in subsiquent requests.

Food Item Insertion 

    {
        "name" : "Oreo Shake",
        "price" : 120.0,
        "food_image" : "https://thumbs.dreamstime.com/b/cookies-cream-milkshake-homemade-tall-glass-39608245.jpg",
        "menu_id": "63049f9c177d4e9aa30644d4"
    }

Create Menu :

    {
        "name" : "Shakes",
        "category" : "cold-drinks"
    }

Update menu :

    {
        "category" : "hot-drinks",
        "start_date" :"2022-09-05T22:16:18Z", 
        "end_date" : "2022-09-06T22:16:18Z"
    }

Insert Table :

    {
        "number_of_guests" : 10,
        "table_number" : 3
    }

Update Table :
    {
        "number_of_guests" : 12
    }

Insert Order :

    {
        "order_date" :"2022-08-23T09:51:34Z",
        "Table_id":  "6304a406177d4e9aa30644d7"
    }

Update Order :

    {
        "table_id": "6304a3eb177d4e9aa30644d6"
    }

Order Item Insert :

    {
        "Table_id" : "6304a406177d4e9aa30644d7",
        "Order_items" : [{
            "quantity" : "S",
            "unit_price" : 120.2,
            "food_id" : "6304a326177d4e9aa30644d5",
            "order_id" : "6304a570177d4e9aa30644d8" 
        }]
    }

Update Order Item :

    {
        "quantity":"L",
        "unit_price":123.23,
        "food_id":"6304a326177d4e9aa30644d5"
    }

Create Invoice :

    {
        "order_id":"6304a570177d4e9aa30644d8",
        "payment_method":"CARD",
        "payment_status" : "PENDING"
    }




