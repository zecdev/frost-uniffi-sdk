import XCTest
@testable import FrostSwift

public class FrostSwiftTests {
    func testFrost() {
        XCTAssertEqual(FrostSwift.frost(), "❄️")
    }
}
