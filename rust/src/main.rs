use std::cell::Cell;
use std::io::{self, Read};
use std::sync::atomic::{AtomicUsize, Ordering};
use std::sync::Mutex;

use actix_web::{
    middleware,
    web::{self, Data},
    App, HttpRequest, HttpResponse, HttpServer,
};

/// simple handle
async fn index(
    counter_mutex: Data<Mutex<usize>>,
    counter_cell: Data<Cell<u32>>,
    counter_atomic: Data<AtomicUsize>,
    req: HttpRequest,
) -> HttpResponse {
    println!("{:?}", req);

    // Increment the counters
    *counter_mutex.lock().unwrap() += 1;
    counter_cell.set(counter_cell.get() + 1);
    counter_atomic.fetch_add(1, Ordering::SeqCst);

    let body = "Hello world!";
    HttpResponse::Ok().body(body)
}

#[actix_web::main]
async fn main() -> io::Result<()> {
    std::env::set_var("RUST_LOG", "actix_web=info");
    env_logger::init();

    // Create some global state prior to building the server
    #[allow(clippy::mutex_atomic)] // it's intentional.
    let counter_mutex = Data::new(Mutex::new(0usize));
    let counter_atomic = Data::new(AtomicUsize::new(0usize));

    // move is necessary to give closure below ownership of counter1
    HttpServer::new(move || {
        // Create some thread-local state
        let counter_cell = Cell::new(0u32);

        App::new()
            .app_data(counter_mutex.clone()) // add shared state
            .app_data(counter_atomic.clone()) // add shared state
            .app_data(Data::new(counter_cell)) // add thread-local state
            .route("/test", web::get().to(test_route))
            // enable logger
            .wrap(middleware::Logger::default())
            // register simple handler
            .service(web::resource("/").to(index))
    })
    .bind(("127.0.0.1", 8080))?
    .run()
    .await
}

async fn test_route(
    req: HttpRequest,
) -> HttpResponse {
    println!("{:?}", req);

    let body = format!(
        "Hello, from test!"
    );
    HttpResponse::Ok().body(body)
}